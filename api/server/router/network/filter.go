package network

import (
	"fmt"

	"github.com/docker/docker/runconfig"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

type filterHandler func([]types.NetworkResource, string) ([]types.NetworkResource, error)

var (
	// AcceptedFilters is an acceptable filters for validation
	AcceptedFilters = map[string]bool{
		"driver": true,
		"type":   true,
		"name":   true,
		"id":     true,
		"label":  true,
	}
)

func filterNetworkByType(nws []types.NetworkResource, netType string) (retNws []types.NetworkResource, err error) {
	switch netType {
	case "builtin":
		for _, nw := range nws {
			if runconfig.IsPreDefinedNetwork(nw.Name) {
				retNws = append(retNws, nw)
			}
		}
	case "custom":
		for _, nw := range nws {
			if !runconfig.IsPreDefinedNetwork(nw.Name) {
				retNws = append(retNws, nw)
			}
		}
	default:
		return nil, fmt.Errorf("Invalid filter: 'type'='%s'", netType)
	}
	return retNws, nil
}

// filterNetworks filters network list according to user specified filter
// and returns user chosen networks
func filterNetworks(nws []types.NetworkResource, filter filters.Args) ([]types.NetworkResource, error) {
	// if filter is empty, return original network list
	if filter.Len() == 0 {
		return nws, nil
	}

	if err := filter.Validate(AcceptedFilters); err != nil {
		return nil, err
	}

	var displayNet []types.NetworkResource
	for _, nw := range nws {
		if filter.Include("driver") {
			if !filter.ExactMatch("driver", nw.Driver) {
				continue
			}
		}
		if filter.Include("name") {
			if !filter.Match("name", nw.Name) {
				continue
			}
		}
		if filter.Include("id") {
			if !filter.Match("id", nw.ID) {
				continue
			}
		}
		if filter.Include("label") {
			if !filter.MatchKVList("label", nw.Labels) {
				continue
			}
		}
		displayNet = append(displayNet, nw)
	}

	if filter.Include("type") {
		var typeNet []types.NetworkResource
		errFilter := filter.WalkValues("type", func(fval string) error {
			passList, err := filterNetworkByType(displayNet, fval)
			if err != nil {
				return err
			}
			typeNet = append(typeNet, passList...)
			return nil
		})
		if errFilter != nil {
			return nil, errFilter
		}
		displayNet = typeNet
	}

	return displayNet, nil
}
