package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

const ipAPI = "https://api.ipify.org/?format=plaintext"

type splitDomain struct {
	sub  string
	root string
}

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func run(doKey, domainString string, updateInterval time.Duration) {
	tokenSource := &tokenSource{
		AccessToken: doKey,
	}
	domain := newSplitDomain(domainString)
	if domain == nil {
		log.Fatalf("Faild to build split domain with: %s", domainString)
	}

	// First update on program start
	update(domain, tokenSource)
	for {
		select {
		case <-time.After(updateInterval):
			update(domain, tokenSource)
		}
	}
}

func update(domain *splitDomain, tokenSource *tokenSource) {
	addr, err := whatsMyIP()
	if err != nil {
		log.Println(err)
		return
	}
	record, err := checkRecord(tokenSource, domain, addr)
	if err != nil {
		log.Println(err)
		return
	}
	if record == nil {
		return
	}
	err = updateRecord(tokenSource, domain, record, addr)
	if err != nil {
		log.Println(err)
		return
	}
}

func updateRecord(tokenSource oauth2.TokenSource, domain *splitDomain,
	record *godo.DomainRecord, addr net.IP) error {
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	editRequest := &godo.DomainRecordEditRequest{
		Data: addr.String(),
	}
	domainRecord, _, err := client.Domains.EditRecord(
		context.TODO(), domain.root, record.ID, editRequest)
	if err != nil {
		return err
	}
	log.Printf("Record updated: %v", domainRecord)
	return nil
}

// checkRecord checks if the A record matches the provided IP
// It if does then nil is returned, otherwise the A record is returned
func checkRecord(tokenSource *tokenSource,
	domain *splitDomain, addr net.IP) (*godo.DomainRecord, error) {
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	records, _, err := client.Domains.Records(context.TODO(), domain.root, nil)
	if err != nil {
		return nil, err
	}

	domainToCheck := domain.root
	if domain.sub != "" {
		domainToCheck = domain.sub
	}
	for _, record := range records {
		if record.Type == "A" && record.Name == domainToCheck {
			recordAddr := net.ParseIP(record.Data)
			if recordAddr == nil {
				return nil, fmt.Errorf("Failed to parse record IP: %s",
					record.Data)
			}
			if !addr.Equal(recordAddr) {
				log.Printf("Record is out of date, domain: %s, record: %v",
					domainToCheck, record)
				return &record, nil
			}
			log.Printf("Record is okay, domain: %s, record: %v",
				domainToCheck, record)
			return nil, nil
		}
	}
	return nil, fmt.Errorf("Couldn't find: %s in %v", domainToCheck, records)
}

func newSplitDomain(domain string) *splitDomain {
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return nil
	}
	if len(domainParts) == 2 {
		return &splitDomain{
			root: domain,
		}
	}
	rootDomainParts := domainParts[len(domainParts)-2:]
	subdomainDomainParts := domainParts[:len(domainParts)-2]
	return &splitDomain{
		root: strings.Join(rootDomainParts, "."),
		sub:  strings.Join(subdomainDomainParts, "."),
	}
}

func whatsMyIP() (net.IP, error) {
	resp, err := http.Get(ipAPI)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status code: %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	addr := net.ParseIP(string(body))
	if addr == nil {
		return nil, fmt.Errorf("Failed to parse address: %v", string(body))
	}
	return addr, nil
}

func main() {
	var (
		updateInterval = flag.Duration("interval", 5*time.Minute,
			"How long to check for updates")
		domainFlag = flag.String("domain", "", "Domain to update")
	)
	flag.Parse()
	doKey := os.Getenv("DO_KEY")
	if doKey == "" {
		log.Fatal("DO_KEY env variable not set")
	}
	if *domainFlag == "" {
		log.Fatal("-domain not passed")
	}

	run(doKey, *domainFlag, *updateInterval)
}
