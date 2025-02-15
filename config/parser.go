package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/xxf098/lite-proxy/common/structure"
	"github.com/xxf098/lite-proxy/outbound"
)

func ParseProxy(mapping map[string]interface{}, namePrefix string) (string, error) {
	decoder := structure.NewDecoder(structure.Option{TagName: "proxy", WeaklyTypedInput: true})
	proxyType, existType := mapping["type"].(string)
	if !existType {
		return "", fmt.Errorf("missing type")
	}

	var (
		err  error
		link string
	)

	switch proxyType {
	case "ss":
		ssOption := &outbound.ShadowSocksOption{}
		err = decoder.Decode(mapping, ssOption)
		if err != nil {
			break
		}
		auth := fmt.Sprintf("%s:%s", ssOption.Cipher, ssOption.Password)
		link = fmt.Sprintf("ss://%s@%s:%d", base64.StdEncoding.EncodeToString([]byte(auth)), ssOption.Server, ssOption.Port)
		if len(ssOption.Name) > 0 {
			link = fmt.Sprintf("%s#%s", link, url.QueryEscape(ssOption.Name))
		}
	case "ssr":
		ssrOption := &outbound.ShadowSocksROption{}
		err = decoder.Decode(mapping, ssrOption)
		if err != nil {
			break
		}
		password := base64.StdEncoding.EncodeToString([]byte(ssrOption.Password))
		link = fmt.Sprintf("%s:%d:%s:%s:%s:%s", ssrOption.Server, ssrOption.Port, ssrOption.Protocol, ssrOption.Cipher, ssrOption.Obfs, password)
		remarks := base64.StdEncoding.EncodeToString([]byte(ssrOption.Name))

		obfsParam := base64.StdEncoding.EncodeToString([]byte(ssrOption.ObfsParam))
		protocolParam := base64.StdEncoding.EncodeToString([]byte(ssrOption.ProtocolParam))
		link = fmt.Sprintf("%s/?obfsparam=%s&remarks=%s&protoparam=%s", link, url.QueryEscape(obfsParam), url.QueryEscape(remarks), url.QueryEscape(protocolParam))
		link = fmt.Sprintf("ssr://%s", base64.StdEncoding.EncodeToString([]byte(link)))
	case "vmess":
		// TODO: h2
		vmessOption := &outbound.VmessOption{
			HTTPOpts: outbound.HTTPOptions{
				Method: "GET",
				Path:   []string{"/"},
			},
		}
		err = decoder.Decode(mapping, vmessOption)
		if err != nil {
			break
		}
		tls := ""
		if vmessOption.TLS {
			tls = "tls"
		}
		host := ""
		if h, ok := vmessOption.WSHeaders["Host"]; ok {
			host = h
		}
		if len(vmessOption.Network) < 1 {
			vmessOption.Network = "tcp"
		}
		skipCertVerify := vmessOption.SkipCertVerify
		if len(vmessOption.ServerName) < 1 {
			skipCertVerify = true
		}
		c := VmessConfigMarshal{
			Ps:             namePrefix + vmessOption.Name,
			Add:            vmessOption.Server,
			Port:           vmessOption.Port,
			Aid:            vmessOption.AlterID,
			ID:             vmessOption.UUID,
			Type:           vmessOption.Cipher,
			TLS:            tls,
			Net:            vmessOption.Network,
			Path:           vmessOption.WSPath,
			Host:           host,
			SkipCertVerify: skipCertVerify,
			ServerName:     vmessOption.ServerName,
		}
		data, err := json.MarshalIndent(&c, "", "    ")
		if err != nil {
			return "", err
		}
		link = fmt.Sprintf("vmess://%s", base64.StdEncoding.EncodeToString(data))
	case "trojan":
		trojanOption := &outbound.TrojanOption{}
		err = decoder.Decode(mapping, trojanOption)
		if err != nil {
			break
		}
		// TODO: SNI
		link = fmt.Sprintf("trojan://%s@%s:%d", trojanOption.Password, trojanOption.Server, trojanOption.Port)
		query := []string{}
		// allowInsecure
		if trojanOption.SkipCertVerify {
			query = append(query, "allowInsecure=1")
		}
		if len(trojanOption.SNI) > 0 {
			query = append(query, fmt.Sprintf("sni=%s", trojanOption.SNI))
		}
		if len(query) > 0 {
			link = fmt.Sprintf("%s?%s", link, strings.Join(query, "&"))
		}
		if len(trojanOption.Remarks) > 0 {
			link = fmt.Sprintf("%s#%s%s", link, namePrefix, url.QueryEscape(trojanOption.Remarks))
		}
		if len(trojanOption.Name) > 0 {
			link = fmt.Sprintf("%s#%s%s", link, namePrefix, url.QueryEscape(trojanOption.Name))
		}
	default:
		return "", fmt.Errorf("unsupport proxy type: %s", proxyType)
	}

	if err != nil {
		return "", err
	}

	return link, nil

}
