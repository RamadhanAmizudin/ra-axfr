package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "net"
    "os"
    "strings"

    "github.com/miekg/dns"
)

type Result struct {
    Domain      string      `json:"domain"`
    Nameservers []string    `json:"nameservers"`
    Results     []DNSRecord `json:"results"`
}

type DNSRecord struct {
    Name  string `json:"name"`
    Type  string `json:"type"`
    Value string `json:"value"`
}

func getNameservers(domain string) ([]string, error) {
    nameservers, err := net.LookupNS(domain)
    if err != nil {
        return nil, err
    }

    var nsList []string
    for _, ns := range nameservers {
        nsList = append(nsList, ns.Host)
    }
    return nsList, nil
}

func performAXFR(domain, ns string) ([]DNSRecord, error) {
    var records []DNSRecord

    transfer := new(dns.Transfer)
    
    m := new(dns.Msg)
    m.SetAxfr(dns.Fqdn(domain))

    nameserver := ns
    if !strings.HasSuffix(nameserver, ".") {
        nameserver += "."
    }

    // Remove trailing dot for net.Dial
    nameserver = nameserver[:len(nameserver)-1]
    conn, err := transfer.In(m, nameserver+":53")
    if err != nil {
        return nil, err
    }

    for r := range conn {
        if r.Error != nil {
            return nil, r.Error
        }
        for _, rr := range r.RR {
            records = append(records, dnsRecordFromRR(rr))
        }
    }
    return records, nil
}

func dnsRecordFromRR(rr dns.RR) DNSRecord {
    header := rr.Header()
    record := DNSRecord{
        Name: header.Name,
        Type: dns.TypeToString[header.Rrtype],
    }

    switch v := rr.(type) {
    case *dns.A:
        record.Value = v.A.String()
    case *dns.AAAA:
        record.Value = v.AAAA.String()
    case *dns.CNAME:
        record.Value = v.Target
    case *dns.MX:
        record.Value = fmt.Sprintf("%d %s", v.Preference, v.Mx)
    case *dns.NS:
        record.Value = v.Ns
    case *dns.PTR:
        record.Value = v.Ptr
    case *dns.SOA:
        record.Value = fmt.Sprintf("%s %s %d %d %d %d %d", v.Ns, v.Mbox, v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl)
    case *dns.SRV:
        record.Value = fmt.Sprintf("%d %d %d %s", v.Priority, v.Weight, v.Port, v.Target)
    case *dns.TXT:
        record.Value = strings.Join(v.Txt, " ")
    default:
        record.Value = rr.String()
    }

    return record
}

func main() {
    domain := flag.String("domain", "", "The domain to look up")
    list := flag.String("list", "", "File containing list of domains to look up")
    outputJSON := flag.Bool("json", false, "Output in JSON format")
    flag.Parse()

    var domains []string

    if *domain != "" {
        domains = append(domains, *domain)
    } else if *list != "" {
        file, err := os.Open(*list)
        if err != nil {
            fmt.Printf("Failed to open list file %s: %v\n", *list, err)
            return
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            domains = append(domains, scanner.Text())
        }
        if err := scanner.Err(); err != nil {
            fmt.Printf("Failed to read list file %s: %v\n", *list, err)
            return
        }
    } else {
        stat, _ := os.Stdin.Stat()
        if (stat.Mode() & os.ModeCharDevice) == 0 {
            scanner := bufio.NewScanner(os.Stdin)
            for scanner.Scan() {
                domains = append(domains, scanner.Text())
            }
            if err := scanner.Err(); err != nil {
                fmt.Printf("Failed to read from stdin: %v\n", err)
                return
            }
        } else {
            fmt.Println("Please provide a domain using -domain, -list, or via stdin")
            return
        }
    }

    results := []Result{}

    for _, domain := range domains {
        nameservers, err := getNameservers(domain)
        if err != nil {
            fmt.Printf("Failed to lookup nameservers for domain %s: %v\n", domain, err)
            continue
        }

        result := Result{
            Domain:      domain,
            Nameservers: nameservers,
        }

        for _, ns := range nameservers {
            records, err := performAXFR(domain, ns)
            if err != nil {
                fmt.Printf("Failed to perform AXFR request to %s: %v\n", ns, err)
                continue
            }
            result.Results = append(result.Results, records...)
        }

        results = append(results, result)
    }

    if *outputJSON {
        output, err := json.Marshal(results)
        if err != nil {
            fmt.Printf("Failed to marshal result to JSON: %v\n", err)
            return
        }
        fmt.Println(string(output))
    } else {
        for _, result := range results {
            fmt.Printf("Domain: %s\n", result.Domain)
            fmt.Println("Nameservers:")
            for _, ns := range result.Nameservers {
                fmt.Printf("  %s\n", ns)
            }
            fmt.Println("AXFR Records:")
            for _, rr := range result.Results {
                fmt.Printf("%s %s %s\n", rr.Name, rr.Type, rr.Value)
            }
        }
    }
}
