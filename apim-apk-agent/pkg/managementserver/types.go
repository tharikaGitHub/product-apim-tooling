/*
 *  Copyright (c) 2024, WSO2 LLC. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package managementserver

// Subscription for struct subscription
type Subscription struct {
	SubStatus     string         `json:"subStatus,omitempty"`
	UUID          string         `json:"uuid,omitempty"`
	Organization  string         `json:"organization,omitempty"`
	SubscribedAPI *SubscribedAPI `json:"subscribedApi,omitempty"`
	TimeStamp     int64          `json:"timeStamp,omitempty"`
}

// SubscriptionList for struct list of applications
type SubscriptionList struct {
	List []Subscription `json:"list"`
}

// SubscribedAPI for struct subscribedAPI
type SubscribedAPI struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// Application for struct application
type Application struct {
	UUID         string            `json:"uuid,omitempty"`
	Name         string            `json:"name,omitempty"`
	Owner        string            `json:"owner,omitempty"`
	Organization string            `json:"organization,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	TimeStamp    int64             `json:"timeStamp,omitempty"`
}

// ApplicationList for struct list of application
type ApplicationList struct {
	List []Application `json:"list"`
}

// ResolvedApplicationList for struct list of resolved application
type ResolvedApplicationList struct {
	List []ResolvedApplication `json:"list"`
}

// ResolvedApplication for struct resolvedApplication
type ResolvedApplication struct {
	UUID            string            `json:"uuid,omitempty"`
	Name            string            `json:"name,omitempty"`
	Owner           string            `json:"owner,omitempty"`
	Organization    string            `json:"organization,omitempty"`
	Attributes      map[string]string `json:"attributes,omitempty"`
	TimeStamp       int64             `json:"timeStamp,omitempty"`
	SecuritySchemes []SecurityScheme  `json:"securitySchemes,omitempty"`
}

// SecurityScheme for struct securityScheme
type SecurityScheme struct {
	SecurityScheme        string `json:"securityScheme,omitempty"`
	ApplicationIdentifier string `json:"applicationIdentifier,omitempty"`
	KeyType               string `json:"keyType,omitempty"`
	EnvID                 string `json:"envID,omitempty"`
}

// ApplicationKeyMapping for struct applicationKeyMapping
type ApplicationKeyMapping struct {
	ApplicationUUID       string `json:"applicationUUID,omitempty"`
	SecurityScheme        string `json:"securityScheme,omitempty"`
	ApplicationIdentifier string `json:"applicationIdentifier,omitempty"`
	KeyType               string `json:"keyType,omitempty"`
	EnvID                 string `json:"envID,omitempty"`
	Timestamp             int64  `json:"timestamp,omitempty"`
	Organization          string `json:"organization,omitempty"`
}

// ApplicationKeyMappingList for struct list of applicationKeyMapping
type ApplicationKeyMappingList struct {
	List []ApplicationKeyMapping `json:"list"`
}

// ApplicationMapping for struct applicationMapping
type ApplicationMapping struct {
	UUID            string `json:"uuid,omitempty"`
	ApplicationRef  string `json:"applicationRef,omitempty"`
	SubscriptionRef string `json:"subscriptionRef,omitempty"`
	Organization    string `json:"organization,omitempty"`
}

// ApplicationMappingList for struct list of applicationMapping
type ApplicationMappingList struct {
	List []ApplicationMapping `json:"list"`
}

// APICPEvent holds data of a specific API event from adapter
type APICPEvent struct {
	Event EventType `json:"event"`
	API   API       `json:"payload"`
}

// EventType is the type of api event. One of (CREATE, UPDATE, DELETE)
type EventType string

const (
	// CreateEvent is create api event
	CreateEvent EventType = "CREATE"
	// DeleteEvent is delete api event
	DeleteEvent EventType = "DELETE"
)

// API holds the api data from adapter api event
type API struct {
	APIUUID          string     `json:"apiUUID"`
	APIName          string     `json:"apiName"`
	APIVersion       string     `json:"apiVersion"`
	IsDefaultVersion bool       `json:"isDefaultVersion"`
	Definition       string     `json:"definition"`
	APIType          string     `json:"apiType"`
	BasePath         string     `json:"basePath"`
	Organization     string     `json:"organization"`
	SystemAPI        bool       `json:"systemAPI"`
	APIProperties    []Property `json:"apiProperties,omitempty"`
	Environment      string     `json:"environment,omitempty"`
	RevisionID       string     `json:"revisionID"`
}

// Property holds the api property
type Property struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
