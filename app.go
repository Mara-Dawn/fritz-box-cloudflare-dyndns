package main

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"net/http"
	"strings"
)

type Parameters struct {
	token   string
	zone    string
	records []string
	ipv4    string
	ipv6    string
}

func health(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func handle_ddns_change(
	req *http.Request,
	done chan<- bool,
	failed chan<- error,
	status chan<- int) {

	parameters, err := parse_params(req)
	if err != nil {
		failed <- err
		status <- http.StatusBadRequest
		done <- true
		return
	}

	if err := apply_changes(parameters); err != nil {
		failed <- err
		status <- http.StatusInternalServerError
		done <- true
		return
	}

	done <- true
}

func apply_changes(parameters *Parameters) error {
	api, err := cloudflare.NewWithAPIToken(parameters.token)
	if err != nil {
		return err
	}

	zone_id, err := api.ZoneIDByName(parameters.zone)
	if err != nil {
		return err
	}

	err = apply_dns_change(api, zone_id, parameters.zone, parameters.ipv4, parameters.ipv6)
	if err != nil {
		fmt.Println("CloudFlare:", err)
	}

	for _, record := range parameters.records {
		url := fmt.Sprintf("%s.%s", record, parameters.zone)
		err = apply_dns_change(api, zone_id, url, parameters.ipv4, parameters.ipv6)
		if err != nil {
			fmt.Println("CloudFlare:", err)
		}
	}

	return nil
}

func apply_dns_change(api *cloudflare.API, zone_id string, url string, ipv4 string, ipv6 string) error {
	fmt.Printf("Checking changes for %s\n", url)

	ctx := context.Background()

	recs, _, err := api.ListDNSRecords(
		ctx,
		cloudflare.ZoneIdentifier(zone_id),
		cloudflare.ListDNSRecordsParams{Name: url},
	)
	if err != nil {
		return err
	}

	var a_record *cloudflare.DNSRecord
	var aaaa_record *cloudflare.DNSRecord

	for _, r := range recs {
		switch r.Type {
		case "A":
			a_record = &r
		case "AAAA":
			aaaa_record = &r
		}
	}

	if ipv4 != "" {
		err := update_record(api, a_record, url, zone_id, "A", ipv4)
		if err != nil {
			return err
		}
	}
	if ipv6 != "" {
		err := update_record(api, aaaa_record, url, zone_id, "AAAA", ipv6)
		if err != nil {
			return err
		}
	}

	return nil
}

func update_record(
	api *cloudflare.API,
	record *cloudflare.DNSRecord,
	url string,
	zone_id string,
	record_type string,
	content string,
) error {
	ctx := context.Background()
	proxied := true

	if record == nil {
		fmt.Printf("  No existing %s record found for %s. Creating new one.\n", record_type, url)
		_, err := api.CreateDNSRecord(
			ctx,
			cloudflare.ZoneIdentifier(zone_id),
			cloudflare.CreateDNSRecordParams{
				Type:    record_type,
				Name:    url,
				Content: content,
				Proxied: &proxied,
			},
		)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("  Existing %s record found for %s.\n", record_type, url)
		if content != record.Content {
			fmt.Printf("  Updating %s record for %s: %s -> %s\n", record_type, url, record.Content, content)
			_, err := api.UpdateDNSRecord(
				ctx,
				cloudflare.ZoneIdentifier(zone_id),
				cloudflare.UpdateDNSRecordParams{
					ID:      record.ID,
					Content: content,
				},
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parse_params(req *http.Request) (*Parameters, error) {
	req.ParseForm()
	token := req.Form.Get("token")
	zone := req.Form.Get("zone")
	record := req.Form.Get("records")
	ipv4 := req.Form.Get("ipv4")
	ipv6 := req.Form.Get("ipv6")

	records := strings.Split(record, ",")

	// fmt.Printf("token: %s\n", token)
	fmt.Printf("zone: %s\n", zone)
	fmt.Printf("records: %s\n", records)
	fmt.Printf("ipv4: %s\n", ipv4)
	fmt.Printf("ipv6: %s\n", ipv6)

	var err_str []string

	if token == "" {
		err_str = append(err_str, "token")
	}
	if zone == "" {
		err_str = append(err_str, "zone")
	}
	if record == "" {
		err_str = append(err_str, "records")
	}
	if ipv4 == "" && ipv6 == "" {
		err_str = append(err_str, "ipv4 or ipv6")
	}

	if len(err_str) > 0 {
		for idx, value := range err_str {
			err_str[idx] = fmt.Sprintf("Missing %s URL parameter.", value)
		}
		err := fmt.Errorf(strings.Join(err_str, "\n"))
		return nil, err
	}

	parameters := &Parameters{token: token, zone: zone, records: records, ipv4: ipv4, ipv6: ipv6}

	return parameters, nil
}

func ddns(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	fmt.Println("\nserver: ddns handler started")
	defer fmt.Println("server: ddns handler ended")

	done := make(chan bool, 1)
	failed := make(chan error, 1)
	status := make(chan int, 1)

	go handle_ddns_change(req, done, failed, status)

	select {
	case <-done:
		select {
		case err := <-failed:
			final_status := <-status
			fmt.Println("server:", err)
			http.Error(w, err.Error(), final_status)
		default:
			fmt.Fprintf(w, "Update successful.")
		}
	case <-ctx.Done():
		err := ctx.Err()
		fmt.Println("server:", err)
		internalError := http.StatusInternalServerError
		http.Error(w, err.Error(), internalError)
	}
}

func main() {

	http.HandleFunc("/health", health)
	http.HandleFunc("/", ddns)

	http.ListenAndServe(":8070", nil)
}
