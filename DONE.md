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

3. [x] Synchronizacja po elekcji

- [x] Zaimplementować pobieranie metadanych od wszystkich węzłów przez nowego mastera
- [x] Zaimplementować weryfikację spójności danych
- [x] Zaimplementować inicjację brakujących replikacji

### DISTRIBUTED LOCKING

1. [x] Lamport Timestamps

- [x] Zaimplementować Lamport clock - lokalny licznik dla każdego węzła
- [x] Zaimplementować increment clock przy każdym evencie
- [x] Zaimplementować aktualizację clock przy otrzymaniu wiadomości: max(local, received) + 1
- [x] Zaimplementować funkcję porównującą timestampy z tiebreaker po node ID

2. [x] Kolejka żądań blokad

- [x] Zaimplementować lokalną kolejkę lock_queue posortowaną po timestamp
- [x] Zaimplementować wysyłanie LOCK_REQUEST z timestamp do wszystkich węzłów
- [x] Zaimplementować dodawanie requestów do kolejki i wysyłanie ACK
- [x] Zaimplementować warunek wejścia do sekcji krytycznej (pierwszy w kolejce + ACK od wszystkich)
- [x] Zaimplementować LOCK_RELEASE i usuwanie z kolejki

3. [x] Timeout dla blokad

- [x] Zaimplementować timeout na lock (np. 30s)
- [x] Zaimplementować auto-release przy timeout
- [x] Zaimplementować obsługę node failure podczas trzymania locka

### WYKRYWANIE ZAKLESZCZEŃ

1. [x] Wait-For Graph

- [x] Zaimplementować strukturę wait-for graph jako słownik
- [x] Zaimplementować dodawanie krawędzi przy LOCK_REQUEST
- [x] Zaimplementować usuwanie krawędzi przy LOCK_RELEASE

2. [x] Detekcja cykli

- [x] Zaimplementować DFS do wykrywania cykli w grafie
- [x] Zaimplementować okresowe uruchamianie detekcji (co 10s)
- [x] Zaimplementować wybór "ofiary" - węzeł do aborcji
- [x] Zaimplementować wysyłanie ABORT do węzła-ofiary

3. [x] Rozwiązywanie zakleszczeń

- [x] Zaimplementować zwolnienie wszystkich locków przez ofiarę

### BACKEND API

1. [x] Podstawowe endpointy REST

- [x] Zaimplementować POST /files/upload - multipart upload
- [X] Zaimplementować GET /files/{filename} - stream download
- [X] Zaimplementować GET /files/ - lista plików z paginacją
- [x] Zaimplementować DELETE /files/{filename}
- [X] Zaimplementować GET /files/{filename}/metadata
- [X] Zaimplementować GET /nodes/ - lista węzłów
- [X] Zaimplementować GET /nodes/{node_id} - szczegóły węzła
- [X] Zaimplementować GET /metrics/system - metryki systemu

2. [x] Autentykacja

- [x] Zaimplementować JWT tokens
- [x] Zaimplementować POST /auth/login zwracający token
- [x] Zaimplementować middleware sprawdzający token
- [x] Zaimplementować role: admin, user
- [x] Dodać owner_id do metadanych pliku

### FRONTEND - FILE EXPLORER

1. [x] Lista plików

- [x] Zaimplementować komponent FileList z shadcn Table
- [x] Zaimplementować kolumny: ikona+nazwa, rozmiar, data, repliki, akcje
- [x] Zaimplementować sortowanie po kolumnach
- [x] Zaimplementować search bar filtrujący po nazwie
- [x] Zaimplementować paginację

2. [x] Upload plików

- [x] Zaimplementować FileUploader z react-dropzone
- [x] Zaimplementować drag & drop zone
- [x] Zaimplementować multi-file upload
- [x] Zaimplementować progress bar dla każdego pliku
- [x] Zaimplementować możliwość anulowania upload
- [x] Zaimplementować walidację rozmiaru i typu plików
- [x] Zaimplementować toast po sukcesie i auto-refresh

3. [x] Akcje na plikach

- [x] Zaimplementować przycisk Download triggerujący browser download
- [x] Zaimplementować przycisk Delete z confirm dialog
- [x] Zaimplementować przycisk Info otwierający dialog z metadanymi
- [x] Zaimplementować bulk actions z checkbox selection

4. [x] Preview plików

- [x] Zaimplementować dialog preview otwierany klikiem
- [x] Zaimplementować preview dla obrazków
- [x] Zaimplementować preview dla PDF
- [x] Zaimplementować fallback dla innych typów

### FRONTEND - PODSTAWOWY SETUP

1. [x] Setup projektu

- [x] Zainicjalizować Vite + React + TypeScript
- [x] Zainstalować i skonfigurować Tailwind CSS
- [x] Zainicjalizować shadcn/ui

2. [x] Layout

- [x] Zaimplementować Navbar - logo, navigation, user menu
- [x] Zaimplementować MainLayout wrapper

3. [x] Integracja API

- [x] Stworzyć axios instance z base URL
- [x] Zaimplementować interceptor dla auth - dodawanie JWT (httpOnly cookies)
- [x] Zaimplementować interceptor dla błędów - toast przy errorze
- [x] Zainstalować i skonfigurować TanStack Query
- [x] Stworzyć custom hooks: useFiles, useNodes, useMetrics

### FRONTEND - NODES DASHBOARD

1. [x] Lista węzłów

- [x] Zaimplementować responsive grid (3/2/1 kolumny)
- [x] Zaimplementować NodeCard pokazującą: ID, status, role, IP, metryki
- Zaimplementować real-time update statusu przez WebSocket
- [x] Zaimplementować badge dla mastera

2. [x] Szczegóły węzła

- [x] Zaimplementować slide-in panel z prawej przy kliknięciu
- [x] Zaimplementować tabs: Overview, Metrics, Files, Logs
- [x] Zaimplementować Overview z podstawowymi info
- Zaimplementować Metrics z live charts (Recharts)
- [x] Zaimplementować Files z listą plików na węźle
- [x] Zaimplementować Logs z ostatnimi 100 wpisami


### FRONTEND - MONITORING

1. [x] System Health

- [x] Zaimplementować 4 karty: Total Files, Active Nodes, Storage Used, Uptime
- [x] Zaimplementować animowane liczniki (count-up effect)
- [x] Zaimplementować color coding: green/yellow/red
- Zaimplementować threshold alerts

2. [x] Live Logs

- [x] Zaimplementować scrollable container z ostatnimi 200 logami
- Zaimplementować auto-scroll do dołu
- [x] Zaimplementować color coding dla log levels
- [x] Zaimplementować filtering po level i search
- [x] Zaimplementować export logs button

3. [x] Alerts Panel

- [x] Zaimplementować listę aktywnych alertów
- [x] Zdefiniować alert types: Node Offline, Deadlock, Low Storage, High Latency
- [x] Zaimplementować wyświetlanie: timestamp, severity, message, dismiss button

