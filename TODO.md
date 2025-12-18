- Dodać pole shards dla partycjonowanych plików
- Zaimplementować rebalancing przy dodaniu/usunięciu węzła

Rebalancing:
- Zaimplementować obliczanie które pliki przenieść przy zmianie topologii
- Zaimplementować tworzenie brakujących replik po offline węzła
- Zaimplementować proces w tle nie blokujący innych operacji
- Zaimplementować tracking progress re-balancingu

Szyfrowanie:
- Zaimplementować szyfrowanie plików AES-256 przy zapisie
- Zaimplementować generowanie master key przy inicjalizacji
- Zaimplementować dystrybucję master key do węzłów
- Zaimplementować deszyfrowanie przy odczycie

Chaos Engineering:
- Zaimplementować scenariusz: losowe zabijanie węzłów co 30s
- Zaimplementować scenariusz: random network delays
- Zaimplementować scenariusz: data corruption
- Przetestować dostępność systemu w chaosie
- Przetestować deadlock detection w chaosie

Snapshot & Recovery:
- Zaimplementować Chandy-Lamport snapshot algorithm
- Dodać button "Create snapshot" w UI
- Zaimplementować zapisywanie snapshotów
- Zaimplementować restore ze snapshota

Kompresja:
- Zaimplementować automatyczną kompresję przy upload
- Zaimplementować dekompresję przy download
- Dodać toggle enable/disable w UI
- Dodać statystyki compression ratio

File versioning:
- Zaimplementować wersjonowanie plików (v1, v2, v3...)
- Dodać historię wersji do UI
- Zaimplementować przywracanie starych wersji
- Zaimplementować vector clocks dla konfliktów