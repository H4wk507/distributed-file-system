## Setup

go:

```bash
wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

go version
```

go migrate:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## TODO:

### PODSTAWOWA ARCHITEKTURA SYSTEMU

1. [x] Węzeł (Node) - podstawowa struktura

- [x] Zaimplementować klasę Node z polami: id, ip, port, role, status
- [x] Zaimplementować metodę heartbeat - wysyłanie sygnału co 5s
- [x] Zaimplementować metodę send_message do komunikacji TCP/UDP
- [x] Zaimplementować metodę receive_message nasłuchującą wiadomości
- [x] Dodać strukturę do trzymania informacji o innych węzłach w klastrze

2. [x] Protokół komunikacji między węzłami

- [x] Zdefiniować format wiadomości (typ, nadawca, odbiorca, payload)
- [x] Zdefiniować typy wiadomości: HEARTBEAT, ELECTION, DATA_REQUEST, LOCK_REQUEST, etc.
- [x] Zaimplementować serializację/deserializację wiadomości
- [x] Zaimplementować retry mechanism przy timeout
- [x] Dodać logowanie wszystkich wiadomości sieciowych

3. [x] Discovery - wykrywanie węzłów

- [x] Zaimplementować bootstrap mechanism - nowy węzeł dostaje listę od seed node
- [x] Zaimplementować broadcast przy dołączeniu nowego węzła
- [x] Zaimplementować usuwanie offline węzłów z listy aktywnych
- [x] Każdy węzeł utrzymuje aktualną listę wszystkich węzłów

### ALGORYTM ELEKCJI LIDERA

1. [x] Bully Algorithm

- [x] Zaimplementować Bully Algorithm - każdy węzeł ma priorytet (wyższy = silniejszy)
- [x] Zaimplementować wysyłanie ELECTION do węzłów o wyższym priorytecie
- [x] Zaimplementować timeout - jeśli nikt nie odpowie = ogłoś siebie masterem
- [x] Zaimplementować broadcast COORDINATOR gdy węzeł zostaje masterem
- [x] Zaimplementować aktualizację informacji o masterze u wszystkich węzłów

2. [x] Wykrywanie awarii mastera

- [x] Zaimplementować detekcję braku heartbeat od mastera (timeout 15s)
- [x] Zaimplementować automatyczne rozpoczęcie elekcji po wykryciu awarii
- [x] Zaimplementować scenariusz powrotu starego mastera (nie przejmuje roli automatycznie)

3. Synchronizacja po elekcji

- Zaimplementować pobieranie metadanych od wszystkich węzłów przez nowego mastera
- Zaimplementować weryfikację spójności danych
- Zaimplementować inicjację brakujących replikacji

### DISTRIBUTED LOCKING

1. Lamport Timestamps

- [x] Zaimplementować Lamport clock - lokalny licznik dla każdego węzła
- [x] Zaimplementować increment clock przy każdym evencie
- [x] Zaimplementować aktualizację clock przy otrzymaniu wiadomości: max(local, received) + 1
- [x] Zaimplementować funkcję porównującą timestampy z tiebreaker po node ID

2. Kolejka żądań blokad

- [x] Zaimplementować lokalną kolejkę lock_queue posortowaną po timestamp
- [x] Zaimplementować wysyłanie LOCK_REQUEST z timestamp do wszystkich węzłów
- [x] Zaimplementować dodawanie requestów do kolejki i wysyłanie ACK
- [x] Zaimplementować warunek wejścia do sekcji krytycznej (pierwszy w kolejce + ACK od wszystkich)
- [x] Zaimplementować LOCK_RELEASE i usuwanie z kolejki

3. Timeout dla blokad

- [x] Zaimplementować timeout na lock (np. 30s)
- [x] Zaimplementować auto-release przy timeout
- [x] Zaimplementować obsługę node failure podczas trzymania locka

### WYKRYWANIE ZAKLESZCZEŃ

1. Wait-For Graph

- [x] Zaimplementować strukturę wait-for graph jako słownik
- [x] Zaimplementować dodawanie krawędzi przy LOCK_REQUEST
- [x] Zaimplementować usuwanie krawędzi przy LOCK_RELEASE
- Zaimplementować agregację informacji o lockach na masterze
- Zaimplementować budowanie globalnego grafu

2. Detekcja cykli

- [x] Zaimplementować DFS do wykrywania cykli w grafie
- [x] Zaimplementować okresowe uruchamianie detekcji (co 10s)
- Zaimplementować wybór "ofiary" - węzeł do aborcji
- Zaimplementować wysyłanie ABORT do węzła-ofiary

3. Rozwiązywanie zakleszczeń

- Zaimplementować zwolnienie wszystkich locków przez ofiarę
- Zaimplementować exponential backoff przed ponowną próbą
- Zaimplementować logowanie incydentów zakleszczenia
- Zaimplementować metrykę liczby zakleszczeń

### PRZECHOWYWANIE I METADANE

1. System metadanych

- Zaimplementować klasę FileMetadata z polami: filename, size, hash, timestamps, owner
- Dodać pole replicas - lista węzłów przechowujących plik
- Dodać pole shards dla partycjonowanych plików
- Zaimplementować globalną bazę metadanych na masterze (SQLite lub in-memory)
- Zaimplementować lokalne metadane na każdym węźle storage

2. Consistent Hashing

- Zaimplementować ring hashing - węzły na pierścieniu 0-2^32
- Zaimplementować funkcję hash_key(filename) -> pozycja na pierścieniu
- Zaimplementować funkcję find_nodes_for_file zwracającą N węzłów storage
- Zaimplementować rebalancing przy dodaniu/usunięciu węzła

3. Lokalne przechowywanie

- Zaimplementować katalog storage dla każdego węzła
- Zaimplementować zapisywanie plików jako {hash}.dat
- Zaimplementować zapisywanie metadanych obok pliku (.metadata.json)
- Zaimplementować indeks lokalny: dict {filename: hash}
- Zaimplementować odczyt plików przy starcie węzła i budowanie indeksu

### REPLIKACJA DANYCH

1. Upload pliku

- Zaimplementować przyjmowanie pliku przez mastera
- Zaimplementować obliczanie hash i sprawdzanie duplikacji
- Zaimplementować wybór N węzłów storage przez consistent hashing
- Zaimplementować wysyłanie pliku równolegle do wszystkich węzłów
- Zaimplementować zbieranie ACK i zapisywanie metadanych
- Zaimplementować wybór náhradního węzła przy braku odpowiedzi

2. Download pliku

- Zaimplementować request do mastera z nazwą pliku
- Zaimplementować lookup metadanych - który węzeł ma plik
- Zaimplementować wybór najbliższego/najmniej obciążonego węzła
- Zaimplementować przekazanie adresu węzła klientowi
- Zaimplementować bezpośredni download klient->storage

3. Wykrywanie niespójności

- Zaimplementować okresowe sprawdzanie replik (raz na godzinę)
- Zaimplementować porównywanie hash replik
- Zaimplementować usuwanie skorumpowanych replik
- Zaimplementować tworzenie nowych replik z dobrego źródła

4. Re-balancing

- Zaimplementować obliczanie które pliki przenieść przy zmianie topologii
- Zaimplementować tworzenie brakujących replik po offline węzła
- Zaimplementować proces w tle nie blokujący innych operacji
- Zaimplementować tracking progress re-balancingu

### BACKEND API

1. Podstawowe endpointy REST

- Zaimplementować POST /files/upload - multipart upload
- Zaimplementować GET /files/{filename} - stream download
- Zaimplementować GET /files/ - lista plików z paginacją
- Zaimplementować DELETE /files/{filename}
- Zaimplementować GET /files/{filename}/metadata
- Zaimplementować GET /nodes/ - lista węzłów
- Zaimplementować GET /nodes/{node_id} - szczegóły węzła
- Zaimplementować GET /metrics/system - metryki systemu

2. WebSocket real-time

- Zaimplementować endpoint WS /ws
- Zdefiniować event types: FILE_UPLOADED, NODE_JOINED, ELECTION_STARTED, DEADLOCK_DETECTED, etc.
- Zaimplementować broadcast eventów do wszystkich klientów
- Zaimplementować connection management

3. Autentykacja

- Zaimplementować JWT tokens
- Zaimplementować POST /auth/login zwracający token
- Zaimplementować middleware sprawdzający token
- Zaimplementować role: admin, user
- Dodać owner_id do metadanych pliku

4. Error handling

- Zdefiniować standardowy format błędu: {error, code, details}
- Zaimplementować odpowiednie HTTP codes: 200, 201, 400, 401, 404, 500, 503
- Zaimplementować timeout handling dla długich operacji
- Zaimplementować retry logic dla operacji rozproszonych

### FRONTEND - PODSTAWOWY SETUP

1. Setup projektu

- Zainicjalizować Vite + React + TypeScript
- Zainstalować i skonfigurować Tailwind CSS
- Zainicjalizować shadcn/ui

2. Layout

- Zaimplementować Navbar - logo, navigation, user menu, dark mode toggle
- Zaimplementować Sidebar - collapsible na mobile
- Zaimplementować MainLayout wrapper
- Zaimplementować Footer

3. Integracja API

- Stworzyć axios instance z base URL
- Zaimplementować interceptor dla auth - dodawanie JWT
- Zaimplementować interceptor dla błędów - toast przy errorze
- Zainstalować i skonfigurować TanStack Query
- Stworzyć custom hooks: useFiles, useNodes, useMetrics

### FRONTEND - FILE EXPLORER

1. Lista plików

- Zaimplementować komponent FileList z shadcn Table
- Zaimplementować kolumny: ikona+nazwa, rozmiar, data, repliki, akcje
- Zaimplementować sortowanie po kolumnach
- Zaimplementować search bar filtrujący po nazwie
- Zaimplementować paginację

2. Upload plików

- Zaimplementować FileUploader z react-dropzone
- Zaimplementować drag & drop zone
- Zaimplementować multi-file upload
- Zaimplementować progress bar dla każdego pliku
- Zaimplementować możliwość anulowania upload
- Zaimplementować walidację rozmiaru i typu plików
- Zaimplementować toast po sukcesie i auto-refresh

3. Akcje na plikach

- Zaimplementować przycisk Download triggerujący browser download
- Zaimplementować przycisk Delete z confirm dialog
- Zaimplementować przycisk Info otwierający dialog z metadanymi
- Zaimplementować context menu (right-click)
- Zaimplementować bulk actions z checkbox selection

4. Preview plików

- Zaimplementować dialog preview otwierany klikiem
- Zaimplementować preview dla obrazków
- Zaimplementować preview dla PDF
- Zaimplementować fallback dla innych typów

### FRONTEND - NODES DASHBOARD

1. Lista węzłów

- Zaimplementować responsive grid (3/2/1 kolumny)
- Zaimplementować NodeCard pokazującą: ID, status, role, IP, metryki
- Zaimplementować real-time update statusu przez WebSocket
- Zaimplementować badge dla mastera

2. Szczegóły węzła

- Zaimplementować slide-in panel z prawej przy kliknięciu
- Zaimplementować tabs: Overview, Metrics, Files, Logs
- Zaimplementować Overview z podstawowymi info
- Zaimplementować Metrics z live charts (Recharts)
- Zaimplementować Files z listą plików na węźle
- Zaimplementować Logs z ostatnimi 100 wpisami

3. Topologia klastra

- Zaimplementować wizualizację grafu używając React Flow lub D3.js
- Zaimplementować węzły jako koła z kolorami statusu
- Zaimplementować krawędzie między węzłami
- Zaimplementować animację podczas elekcji
- Zaimplementować interaktywność - hover tooltip, klik otwiera detail

4. Election History

- Zaimplementować timeline przeszłych elekcji
- Zaimplementować wyświetlanie: timestamp, trigger, candidates, winner
- Zaimplementować filtrowanie po dacie
- Zaimplementować export do CSV

### FRONTEND - MONITORING

1. System Health

- Zaimplementować 4 karty: Total Files, Active Nodes, Storage Used, Uptime
- Zaimplementować animowane liczniki (count-up effect)
- Zaimplementować color coding: green/yellow/red
- Zaimplementować threshold alerts

2. Performance Charts

- Zaimplementować line chart: throughput (files/min) za 24h
- Zaimplementować line chart: latency za 24h
- Zaimplementować bar chart: storage per node
- Zaimplementować pie chart: file types breakdown
- Zaimplementować auto-update co 10s przez WebSocket

3. Live Logs

- Zaimplementować scrollable container z ostatnimi 200 logami
- Zaimplementować auto-scroll do dołu
- Zaimplementować color coding dla log levels
- Zaimplementować filtering po level i search
- Zaimplementować export logs button

4. Alerts Panel

- Zaimplementować listę aktywnych alertów
- Zdefiniować alert types: Node Offline, Deadlock, Low Storage, High Latency
- Zaimplementować wyświetlanie: timestamp, severity, message, dismiss button
- Zaimplementować toast przy nowym alercie
- Zaimplementować archiwum rozwiązanych alertów

### BEZPIECZEŃSTWO

1. Szyfrowanie danych

- Zaimplementować szyfrowanie plików AES-256 przy zapisie
- Zaimplementować generowanie master key przy inicjalizacji
- Zaimplementować dystrybucję master key do węzłów
- Zaimplementować deszyfrowanie przy odczycie

2. Access Control Lists

- Dodać do metadanych: permissions {owner, read[], write[]}
- Zaimplementować sprawdzanie uprawnień przy każdej operacji
- Zaimplementować możliwość modyfikacji permissions przez ownera
- Zaimplementować UI dialog "Share file"

3. Audit Log

- Zaimplementować logowanie każdej operacji: {timestamp, user, action, resource, result, ip}
- Zaimplementować append-only storage dla audit logs
- Zaimplementować UI dla przeglądania audit logs
- Zaimplementować filtering i export
- Zaimplementować retention policy (90 dni)

### TESTOWANIE

1. Testy jednostkowe

- Napisać testy dla consistent hashing
- Napisać testy dla Lamport clock
- Napisać testy dla wait-for graph (detekcja cykli)
- Napisać testy dla file hashing
- Napisać testy dla metadata serialization

2. Testy integracyjne

- Napisać test: upload -> download -> verify checksum
- Napisać test: upload -> delete -> download = 404
- Napisać test: upload do 3 węzłów -> plik na wszystkich
- Napisać test: zabicie węzła storage -> download z innej repliki działa
- Napisać test: zabicie mastera -> elekcja -> operacje działają
- Napisać test: concurrent lock requests -> tylko jeden dostaje

3. Testy mutacyjne

- Uruchomić mutmut na kluczowych modułach
- Osiągnąć mutation score >80%
- Przeanalizować survived mutants
- Dodać brakujące testy

4. Chaos Engineering

- Zaimplementować scenariusz: losowe zabijanie węzłów co 30s
- Zaimplementować scenariusz: random network delays
- Zaimplementować scenariusz: data corruption
- Przetestować dostępność systemu w chaosie
- Przetestować deadlock detection w chaosie

5. Testy obciążeniowe

- Napisać scenariusz locust: 100 concurrent uploads
- Zmierzyć throughput, latency (p50, p95, p99), error rate
- Napisać scenariusz: 1000 concurrent downloads
- Napisać scenariusz: mixed workload (70% read, 30% write)
- Zidentyfikować bottlenecki
- Target: >100 req/sec, <500ms latency

### DOKUMENTACJA (LaTeX)

1. Model systemu

- Napisać definicję rozproszonego systemu plików
- Napisać uzasadnienie wyboru tego typu systemu
- Narysować diagram architektury master-storage-client
- Określić założenia systemu

2. Model matematyczny

- Zdefiniować przestrzeń stanów S
- Zdefiniować funkcje przejścia między stanami
- Wyprowadzić wzór consistent hashing
- Zdefiniować formalnie Lamport timestamps
- Zdefiniować wait-for graph matematycznie

3. Algorytmy - opis szczegółowy

- Napisać pseudokod Bully Algorithm + proof + złożoność
- Napisać pseudokod Lamport Mutual Exclusion + proof + złożoność
- Napisać pseudokod Deadlock Detection (DFS) + złożoność
- Napisać pseudokod Consistent Hashing + rebalancing + złożoność

4. Wzorce projektowe

- Zidentyfikować użyte wzorce: Singleton, Observer, Strategy, Factory, Command
- Opisać gdzie i dlaczego użyto każdego wzorca
- Narysować diagramy UML dla wzorców

5. Analiza bezpieczeństwa

- Zdefiniować threat model (możliwe ataki)
- Opisać mitigation strategies
- Przeprowadzić vulnerability testing
- Stworzyć security audit checklist

6. Diagramy UML (minimum 5)

- Narysować Class Diagram - klasy i relacje
- Narysować Sequence Diagram - proces elekcji
- Narysować Sequence Diagram - upload z replikacją
- Narysować State Diagram - stany węzła
- Narysować Activity Diagram - detekcja zakleszczenia
- Narysować Component Diagram - moduły systemu
- Narysować Deployment Diagram - rozmieszczenie

7. Schemat bazy danych

- Zdefiniować tabelę Files
- Zdefiniować tabelę Replicas
- Zdefiniować tabelę Nodes
- Zdefiniować tabelę Locks
- Narysować ER diagram z relacjami

8. Opis fragmentów kodu

- Wybrać 4-5 kluczowych fragmentów kodu
- Wyjaśnić szczegółowo co robi każda linia
- Wyjaśnić dlaczego wybrano takie rozwiązanie
- Opisać alternatywne podejścia

9. Analiza złożoności

- Obliczyć złożoność upload pliku
- Obliczyć złożoność download pliku
- Obliczyć złożoność elekcji
- Obliczyć złożoność deadlock detection
- Obliczyć złożoność search file

10. Raport z testów

- Stworzyć tabelę: test case, input, expected, actual, status
- Dodać wykresy code coverage per moduł
- Dodać wyniki mutation testing
- Dodać wykresy z load testingu
- Napisać wnioski i możliwe usprawnienia

11. Instrukcja wdrożenia

- Określić wymagania systemowe
- Napisać krok po kroku instalację
- Opisać konfigurację (config files)
- Opisać uruchomienie (docker-compose lub manual)
- Dodać sekcję troubleshooting

### ZAAWANSOWANE FEATURES (OPCJONALNE)

1. Snapshot & Recovery

- Zaimplementować Chandy-Lamport snapshot algorithm
- Dodać button "Create snapshot" w UI
- Zaimplementować zapisywanie snapshotów
- Zaimplementować restore ze snapshota

2. Deduplikacja

- Zaimplementować content-based deduplication po hash
- Zaimplementować reference counting
- Zaimplementować garbage collection
- Dodać statystyki oszczędności miejsca do UI

3. Kompresja

- Zaimplementować automatyczną kompresję przy upload
- Zaimplementować dekompresję przy download
- Dodać toggle enable/disable w UI
- Dodać statystyki compression ratio

4. File versioning

- Zaimplementować wersjonowanie plików (v1, v2, v3...)
- Dodać historię wersji do UI
- Zaimplementować przywracanie starych wersji
- Zaimplementować vector clocks dla konfliktów

5. Geo-replication

- Zaimplementować klastry w różnych "regionach"
- Zaimplementować cross-region replication
- Zaimplementować conflict resolution (last-write-wins lub vector clocks)
- Dodać wybór regionu do UI
