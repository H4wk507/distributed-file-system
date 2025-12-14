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

[x] W teorii powinno dać sie odpalić web, api, db, i kazdy node na osobnych maszynach, komunikacja przez TCP

### PRZECHOWYWANIE I METADANE

1. System metadanych

- [x] Zaimplementować klasę FileMetadata z polami: filename, size, hash, timestamps, owner
- [x] Dodać pole replicas - lista węzłów przechowujących plik
- Dodać pole shards dla partycjonowanych plików
- [x] Zaimplementować globalną bazę metadanych na masterze (SQLite lub in-memory)
- [x] Zaimplementować lokalne metadane na każdym węźle storage

2. Consistent Hashing

- [x] Zaimplementować ring hashing - węzły na pierścieniu 0-2^32
- [x] Zaimplementować funkcję hash_key(filename) -> pozycja na pierścieniu
- [x] Zaimplementować funkcję find_nodes_for_file zwracającą N węzłów storage
- Zaimplementować rebalancing przy dodaniu/usunięciu węzła

3. [x] Lokalne przechowywanie

- [x] Zaimplementować katalog storage dla każdego węzła
- [x] Zaimplementować zapisywanie plików jako {hash}.dat
- [x] Zaimplementować zapisywanie metadanych obok pliku (.metadata.json)
- [x] Zaimplementować indeks lokalny: dict {filename: hash}
- [x] Zaimplementować odczyt plików przy starcie węzła i budowanie indeksu

### REPLIKACJA DANYCH

1. Upload pliku

- [x] Zaimplementować przyjmowanie pliku przez mastera
- Zaimplementować obliczanie hash i sprawdzanie duplikacji
- [x] Zaimplementować wybór N węzłów storage przez consistent hashing
- [x] Zaimplementować wysyłanie pliku równolegle do wszystkich węzłów
- [x] Zaimplementować zbieranie ACK i zapisywanie metadanych
- Zaimplementować wybór następnego węzła przy braku odpowiedzi

2. Download pliku

- [x] Zaimplementować request do mastera z nazwą pliku
- [x] Zaimplementować lookup metadanych - który węzeł ma plik
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

- [x] Zaimplementować POST /files/upload - multipart upload
- [X] Zaimplementować GET /files/{filename} - stream download
- [X] Zaimplementować GET /files/ - lista plików z paginacją
- [x] Zaimplementować DELETE /files/{filename}
- [X] Zaimplementować GET /files/{filename}/metadata
- Zaimplementować GET /nodes/ - lista węzłów
- Zaimplementować GET /nodes/{node_id} - szczegóły węzła
- Zaimplementować GET /metrics/system - metryki systemu

2. WebSocket real-time

- Zaimplementować endpoint WS /ws
- Zdefiniować event types: FILE_UPLOADED, NODE_JOINED, ELECTION_STARTED, DEADLOCK_DETECTED, etc.
- Zaimplementować broadcast eventów do wszystkich klientów
- Zaimplementować connection management

3. [x] Autentykacja

- [x] Zaimplementować JWT tokens
- [x] Zaimplementować POST /auth/login zwracający token
- [x] Zaimplementować middleware sprawdzający token
- [x] Zaimplementować role: admin, user
- [x] Dodać owner_id do metadanych pliku

4. Error handling

- Zdefiniować standardowy format błędu: {error, code, details}
- Zaimplementować odpowiednie HTTP codes: 200, 201, 400, 401, 404, 500, 503
- Zaimplementować timeout handling dla długich operacji
- Zaimplementować retry logic dla operacji rozproszonych

### BEZPIECZEŃSTWO

1. Szyfrowanie danych

- Zaimplementować szyfrowanie plików AES-256 przy zapisie
- Zaimplementować generowanie master key przy inicjalizacji
- Zaimplementować dystrybucję master key do węzłów
- Zaimplementować deszyfrowanie przy odczycie

### TESTOWANIE

1. Testy jednostkowe

- [x] Napisać testy dla consistent hashing
- [x] Napisać testy dla wait-for graph (detekcja cykli)
- [x] Napisać testy dla file hashing

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
