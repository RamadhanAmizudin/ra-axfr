# ra-axfr
Basic tool to extract perform AXFR DNS Request

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
