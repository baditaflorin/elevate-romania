package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// OSMAPIClient handles OSM API operations
type OSMAPIClient struct {
	client *http.Client
	dryRun bool
}

// OSMNode represents a node element in OSM XML
type OSMNode struct {
	XMLName   xml.Name  `xml:"osm"`
	Version   string    `xml:"version,attr"`
	Generator string    `xml:"generator,attr"`
	Node      *NodeData `xml:"node,omitempty"`
}

// NodeData contains node information
type NodeData struct {
	ID        int64     `xml:"id,attr"`
	Version   int       `xml:"version,attr"`
	Changeset int       `xml:"changeset,attr"`
	Lat       float64   `xml:"lat,attr"`
	Lon       float64   `xml:"lon,attr"`
	Tags      []NodeTag `xml:"tag"`
}

// NodeTag represents a tag on a node
type NodeTag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

// OSMWay represents a way element in OSM XML
type OSMWay struct {
	XMLName   xml.Name `xml:"osm"`
	Version   string   `xml:"version,attr"`
	Generator string   `xml:"generator,attr"`
	Way       *WayData `xml:"way,omitempty"`
}

// WayData contains way information
type WayData struct {
	ID        int64     `xml:"id,attr"`
	Version   int       `xml:"version,attr"`
	Changeset int       `xml:"changeset,attr"`
	Tags      []NodeTag `xml:"tag"`
	Nodes     []WayNode `xml:"nd"`
}

// WayNode represents a node reference in a way
type WayNode struct {
	Ref int64 `xml:"ref,attr"`
}

// NewOSMAPIClient creates a new OSM API client
func NewOSMAPIClient(client *http.Client, dryRun bool) *OSMAPIClient {
	return &OSMAPIClient{
		client: client,
		dryRun: dryRun,
	}
}

// FetchNode fetches a node from OSM
func (api *OSMAPIClient) FetchNode(nodeID int64) (*NodeData, error) {
	url := fmt.Sprintf("https://api.openstreetmap.org/api/0.6/node/%d", nodeID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch node: status code %d: %s", resp.StatusCode, string(body))
	}

	var osmNode OSMNode
	if err := xml.NewDecoder(resp.Body).Decode(&osmNode); err != nil {
		return nil, fmt.Errorf("failed to decode node XML: %v", err)
	}

	if osmNode.Node == nil {
		return nil, fmt.Errorf("no node data in response")
	}

	return osmNode.Node, nil
}

// FetchWay fetches a way from OSM
func (api *OSMAPIClient) FetchWay(wayID int64) (*WayData, error) {
	url := fmt.Sprintf("https://api.openstreetmap.org/api/0.6/way/%d", wayID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch way: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch way: status code %d: %s", resp.StatusCode, string(body))
	}

	var osmWay OSMWay
	if err := xml.NewDecoder(resp.Body).Decode(&osmWay); err != nil {
		return nil, fmt.Errorf("failed to decode way XML: %v", err)
	}

	if osmWay.Way == nil {
		return nil, fmt.Errorf("no way data in response")
	}

	return osmWay.Way, nil
}

// UpdateNode updates a node in OSM
func (api *OSMAPIClient) UpdateNode(node *NodeData, changesetID int) error {
	if api.dryRun {
		return nil
	}

	// Set changeset ID
	node.Changeset = changesetID

	osmNode := OSMNode{
		Version:   "0.6",
		Generator: "elevate-romania",
		Node:      node,
	}

	xmlData, err := xml.MarshalIndent(osmNode, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal node XML: %v", err)
	}

	url := fmt.Sprintf("https://api.openstreetmap.org/api/0.6/node/%d", node.ID)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(xmlData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update node: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update node: status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateWay updates a way in OSM
func (api *OSMAPIClient) UpdateWay(way *WayData, changesetID int) error {
	if api.dryRun {
		return nil
	}

	// Set changeset ID
	way.Changeset = changesetID

	osmWay := OSMWay{
		Version:   "0.6",
		Generator: "elevate-romania",
		Way:       way,
	}

	xmlData, err := xml.MarshalIndent(osmWay, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal way XML: %v", err)
	}

	url := fmt.Sprintf("https://api.openstreetmap.org/api/0.6/way/%d", way.ID)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(xmlData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update way: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update way: status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// MergeTags merges new tags with existing tags, updating values for existing keys
func MergeTags(existingTags []NodeTag, newTags map[string]string) []NodeTag {
	// Create a map of existing tags
	tagMap := make(map[string]string)
	for _, tag := range existingTags {
		tagMap[tag.Key] = tag.Value
	}

	// Update with new tags
	for key, value := range newTags {
		tagMap[key] = value
	}

	// Convert back to slice
	var result []NodeTag
	for key, value := range tagMap {
		result = append(result, NodeTag{
			Key:   key,
			Value: value,
		})
	}

	return result
}
