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