package main

type NamespaceList struct {
	Items []Namespace `json:"items"`
}

type Namespace struct {
	Metadata Metadata `json:"metadata"`
}

type ConfigMapList struct {
	Items []ConfigMap `json:"items"`
}

type ConfigMap struct {
	ApiVersion string            `json:"apiVersion"`
	Data       map[string]string `json:"data"`
	Kind       string            `json:"kind"`
	Metadata   Metadata          `json:"metadata"`
}

type SecretList struct {
	Items []Secret `json:"items"`
}

type Secret struct {
	ApiVersion string            `json:"apiVersion"`
	Data       map[string]string `json:"data"`
	Kind       string            `json:"kind"`
	Metadata   Metadata          `json:"metadata"`
	Type       string            `json:"type"`
}

type Metadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
