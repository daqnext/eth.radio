package ensc

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	lru "github.com/hashicorp/golang-lru"
	"github.com/labstack/gommon/log"
	ens "github.com/wealdtech/go-ens/v3"
)

var emptyContentHash = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

// ENS is a plugin that returns information held in the Ethereum Name Service.
type ENS struct {
	Client   *ethclient.Client
	Registry *ens.Registry
}

func (e ENS) HasRecords(domain string, name string) (bool, error) {
	// See if this has a contenthash record.
	resolver, err := e.getResolver(domain)
	if err != nil {
		return false, err
	}
	bytes, err := resolver.Contenthash()
	if err == nil && len(bytes) > 0 {
		return true, err
	}

	// See if this has DNS records.
	dnsResolver, err := e.getDNSResolver(strings.TrimSuffix(domain, "."))
	if err != nil {
		return false, err
	}
	return dnsResolver.HasRecords(name)
}

// Query queries a given domain/name/resource combination
func (e ENS) Query(domain string, name string) (string, error) {
	log.Debugf("request for name %s in domain %v", name, domain)

	// If the requested domain has a content hash we alter a number of the records returned
	var contentHash []byte
	hasContentHash := false
	var err error

	contentHash, err = e.obtainContentHash(domain)
	hasContentHash = err == nil && bytes.Compare(contentHash, emptyContentHash) > 0

	if hasContentHash {

		// Also provide dnslink for compatibility with older IPFS gateways
		contentHashStr, err := ens.ContenthashToString(contentHash)
		if err != nil {
			return "", err
		}

		// for debug
		// contentHashStr = fmt.Sprintf("%s 3600 IN TXT \"contenthash=0x%x\"", name, contentHash)
		// contentHashStr = fmt.Sprintf("%s 3600 IN TXT \"dnslink=%s\"", name, contentHashStr)
		return contentHashStr, nil
	}

	return "", errors.New(fmt.Sprintf("%s contentHash not found", name))
}

func (e ENS) obtainContentHash(domain string) ([]byte, error) {
	ethDomain := strings.TrimSuffix(domain, ".")
	resolver, err := e.getResolver(ethDomain)
	if err != nil {
		return []byte{}, nil
	}

	return resolver.Contenthash()
}

func (e *ENS) getDNSResolver(domain string) (*ens.DNSResolver, error) {
	if !dnsResolverCache.Contains(domain) {
		resolver, err := ens.NewDNSResolver(e.Client, domain)
		if err == nil {
			dnsResolverCache.Add(domain, resolver)
		} else {
			if err.Error() == "no contract code at given address" ||
				strings.HasSuffix(err.Error(), " is not a DNS resolver contract") {
				dnsResolverCache.Add(domain, nil)
			}
		}
	}
	resolver, _ := dnsResolverCache.Get(domain)
	if resolver == nil {
		return nil, errors.New("no resolver")
	}
	return resolver.(*ens.DNSResolver), nil
}

func (e *ENS) newDNSResolver(domain string) (*ens.DNSResolver, error) {
	// Obtain the resolver address for this domain
	resolver, err := e.Registry.ResolverAddress(domain)
	if err != nil {
		return nil, err
	}
	return ens.NewDNSResolverAt(e.Client, domain, resolver)
}

func (e *ENS) getResolver(domain string) (*ens.Resolver, error) {
	if !resolverCache.Contains(domain) {
		resolver, err := e.newResolver(domain)
		if err == nil {
			resolverCache.Add(domain, resolver)
		} else {
			if err.Error() == "no contract code at given address" ||
				strings.HasSuffix(err.Error(), " is not a resolver contract") {
				resolverCache.Add(domain, nil)
			}
		}
	}
	resolver, _ := resolverCache.Get(domain)
	if resolver == nil {
		return nil, errors.New("no resolver")
	}
	return resolver.(*ens.Resolver), nil
}

func (e *ENS) newResolver(domain string) (*ens.Resolver, error) {
	// Obtain the resolver address for this domain
	resolver, err := e.Registry.ResolverAddress(domain)
	if err != nil {
		return nil, err
	}
	return ens.NewResolverAt(e.Client, domain, resolver)
}

var resolverCache *lru.Cache
var dnsResolverCache *lru.Cache

func init() {
	resolverCache, _ = lru.New(16)
	dnsResolverCache, _ = lru.New(16)
}
