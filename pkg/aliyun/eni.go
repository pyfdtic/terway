package aliyun

import (
	"fmt"
	"net"

	apiErr "github.com/AliyunContainerService/terway/pkg/aliyun/errors"
	"github.com/AliyunContainerService/terway/pkg/aliyun/metadata"
	"github.com/AliyunContainerService/terway/types"

	"github.com/pkg/errors"
)

const maxSinglePageSize = 100

// ENIInfoGetter interface to get eni information
type ENIInfoGetter interface {
	GetENIConfigByMac(mac string) (*types.ENI, error)
	GetENIPrivateAddressesByMAC(mac string) ([]net.IP, error)
	GetENIPrivateIPv6AddressesByMAC(mac string) ([]net.IP, error)
	GetAttachedENIs(containsMainENI bool) ([]*types.ENI, error)
	GetSecondaryENIMACs() ([]string, error)
}

type ENIMetadata struct {
	ignoreLinkNotExist bool
	ipFamily           *types.IPFamily
}

func NewENIMetadata(ignoreLinkNotExist bool, ipFamily *types.IPFamily) *ENIMetadata {
	return &ENIMetadata{
		ignoreLinkNotExist: ignoreLinkNotExist,
		ipFamily:           ipFamily,
	}
}

func (e *ENIMetadata) GetENIConfigByMac(mac string) (*types.ENI, error) {
	eni := types.ENI{
		MAC: mac,
	}
	var err error

	eni.ID, err = metadata.GetENIID(mac)
	if err != nil {
		return nil, fmt.Errorf("error get eni by id %s, %w", eni.ID, err)
	}

	ip, err := metadata.GetENIPrimaryIP(mac)
	if err != nil {
		return nil, fmt.Errorf("error get eni primary ip by %s, %w", mac, err)
	}

	eni.PrimaryIP = types.IPSet{
		IPv4: ip,
	}
	gw, err := metadata.GetENIGateway(mac)
	if err != nil {
		return nil, fmt.Errorf("error get eni gateway by mac %s, %w", mac, err)
	}

	vSwitchCIDR, err := metadata.GetVSwitchCIDR(mac)
	if err != nil {
		return nil, fmt.Errorf("error get eni vSwitchCIDR from metaserver, mac: %s, %w", mac, err)
	}

	var v6gw net.IP
	var vSwitchIPv6CIDR *net.IPNet
	if e.ipFamily.IPv6 {
		v6gw, err = metadata.GetENIV6Gateway(mac)
		if err != nil {
			return nil, fmt.Errorf("error get eni ipv6 gateway from metaserver, mac: %s, %w", mac, err)
		}

		vSwitchIPv6CIDR, err = metadata.GetVSwitchIPv6CIDR(mac)
		if err != nil {
			return nil, fmt.Errorf("error get eni vSwitchIPv6CIDR from metaserver, mac: %s, %w", mac, err)
		}
	}

	eni.VSwitchCIDR = types.IPNetSet{
		IPv4: vSwitchCIDR,
		IPv6: vSwitchIPv6CIDR,
	}

	eni.GatewayIP = types.IPSet{
		IPv4: gw,
		IPv6: v6gw,
	}

	vswitch, err := metadata.GetENIVSwitchID(mac)
	if err != nil {
		return nil, errors.Wrapf(err, "error get eni vswitch from metaserver, mac: %s", mac)
	}
	eni.VSwitch = vswitch

	return &eni, nil
}

func (e *ENIMetadata) GetENIConfigByID(eniID string) (*types.ENI, error) {
	macs, err := metadata.GetENIsMAC()
	if err != nil {
		return nil, err
	}
	for _, mac := range macs {
		id, err := metadata.GetENIID(mac)
		if err != nil {
			return nil, errors.Wrapf(err, "error get eni id for mac: %s from metadata", mac)
		}
		if eniID == id {
			return e.GetENIConfigByMac(mac)
		}
	}
	return nil, errors.Wrapf(apiErr.ErrNotFound, fmt.Sprintf("eni id: %s", eniID))
}

func (e *ENIMetadata) GetENIPrivateAddressesByMAC(mac string) ([]net.IP, error) {
	return metadata.GetENIPrivateIPs(mac)
}

func (e *ENIMetadata) GetENIPrivateIPv6AddressesByMAC(mac string) ([]net.IP, error) {
	return metadata.GetENIPrivateIPv6IPs(mac)
}

func (e *ENIMetadata) GetAttachedENIs(containsMainENI bool) ([]*types.ENI, error) {
	var enis []*types.ENI

	mainENIMac := GetInstanceMeta().PrimaryMAC

	macs, err := metadata.GetENIsMAC()
	if err != nil {
		return nil, err
	}
	for _, mac := range macs {
		if !containsMainENI {
			if mac == mainENIMac {
				continue
			}
		}
		eni, err := e.GetENIConfigByMac(mac)
		if err != nil {
			return nil, errors.Wrapf(err, "error get eni info for mac: %s from metadata", mac)
		}
		enis = append(enis, eni)
	}
	return enis, nil
}

// GetSecondaryENIMACs return secondary ENI macs
func (e *ENIMetadata) GetSecondaryENIMACs() ([]string, error) {
	var result []string

	mainENIMac := GetInstanceMeta().PrimaryMAC

	macs, err := metadata.GetENIsMAC()
	if err != nil {
		return nil, err
	}
	for _, mac := range macs {
		if mac == mainENIMac {
			continue
		}
		result = append(result, mac)
	}
	return result, nil
}
