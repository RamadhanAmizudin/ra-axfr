# ra-axfr
Basic tool to perform AXFR DNS request

### Installation

```
go install github.com/RamadhanAmizudin/ra-axfr@latest
```

### Usage for single domain
```
ra-axfr -domain <domain>
```

### Usage via list of file
```
ra-axfr -list <domains.txt>
```

### Usage via STDIN
```
cat domains.txt | ra-axfr
```

Options:
```
-domain      Single Domain
-list 	     File containing list of domains to look up
-json 	     Output as json
```

Example:
```
ra-axfr -list domains.txt -json
```

## Status
This project was purely for my personal learning. If it isn't obvious, this shouldn't be incorporated in any type of application, and the only reason it is open source is that if someone would find useful information or parts from it.
