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

1. [x] Upload pliku

- [x] Zaimplementować przyjmowanie pliku przez mastera
- [x] Zaimplementować wybór N węzłów storage przez consistent hashing
- [x] Zaimplementować wysyłanie pliku równolegle do wszystkich węzłów
- [x] Zaimplementować zbieranie ACK i zapisywanie metadanych

2. [x] Download pliku

- [x] Zaimplementować request do mastera z nazwą pliku
- [x] Zaimplementować lookup metadanych - który węzeł ma plik

3. Re-balancing

- Zaimplementować obliczanie które pliki przenieść przy zmianie topologii
- Zaimplementować tworzenie brakujących replik po offline węzła
- Zaimplementować proces w tle nie blokujący innych operacji
- Zaimplementować tracking progress re-balancingu

### BEZPIECZEŃSTWO

1. Szyfrowanie danych

- Zaimplementować szyfrowanie plików AES-256 przy zapisie
- Zaimplementować generowanie master key przy inicjalizacji
- Zaimplementować dystrybucję master key do węzłów
- Zaimplementować deszyfrowanie przy odczycie

### TESTOWANIE

1. [x] Testy jednostkowe

- [x] Napisać testy dla consistent hashing
- [x] Napisać testy dla wait-for graph (detekcja cykli)
- [x] Napisać testy dla file hashing

2. Chaos Engineering

- Zaimplementować scenariusz: losowe zabijanie węzłów co 30s
- Zaimplementować scenariusz: random network delays
- Zaimplementować scenariusz: data corruption
- Przetestować dostępność systemu w chaosie
- Przetestować deadlock detection w chaosie

3. Testy obciążeniowe

- Napisać scenariusz locust: 100 concurrent uploads
- Zmierzyć throughput, latency (p50, p95, p99), error rate
- Napisać scenariusz: 1000 concurrent downloads
- Napisać scenariusz: mixed workload (70% read, 30% write)
- Zidentyfikować bottlenecki
- Target: >100 req/sec, <500ms latency

### DOKUMENTACJA

1. Architektura systemu

- Narysować diagram architektury: Master Node ↔ Storage Nodes ↔ Client (Web UI)
- Opisać rolę każdego komponentu (master koordynuje, storage przechowuje, client wyświetla)
- Opisać przepływ danych przy upload/download

2. Schemat bazy danych

- Zdefiniować tabele: Files, Replicas
- Narysować ER diagram z relacjami

3. Kluczowe algorytmy (krótki opis)

- Consistent Hashing - jak wybieramy węzły do przechowywania pliku
- Bully Algorithm - jak wybieramy nowego mastera przy awarii
- Deadlock Detection - jak wykrywamy zakleszczenia (wait-for graph + DFS)
- Distributed Locking - jak zarządzamy blokadami w systemie rozproszonym
- Streaming, Protokół - jak przesyłamy pliki między węzłami
- Chunked upload - jak przesyłamy pliki w partiach

4. Instrukcja uruchomienia

- Wymagania: Docker, Go 1.25+, Node.js 22+
- Uruchomienie: `docker-compose up`
- Konfiguracja portów i zmiennych środowiskowych

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