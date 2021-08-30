//nolint
package main

import (
	"fmt"
	"log"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
	"gopkg.in/yaml.v3"
)

const code = `
apiVersion: apps/v1
kind: Deployment
metadata:
  #+operator-builder:collection:field:name=webstoreName,type=string
  name: webstore-deploy
spec:
  #+operator-builder:field:name=webStoreReplicas,default=2,type=int,description="Hello World"
  replicas: 2 
  selector:
    #+potato:pancakes
    matchLabels:
      app: webstore
  template:
    metadata:
      labels:
        # +operator-builder:field:name=webstoreAppLabel,type=string,default=webstore
        app: webstore
    spec:
      containers:
      - name: webstore-container
        #+operator-builder:field:name=webStoreImage,type=string,description="Defines the web store image"
        image: nginx:1.17
        ports:
        - containerPort: 8080
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: webstore-ing
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: app.acme.com
    http:
      paths:
      - path: /
        backend:
          serviceName: webstorep-svc
          servicePort: 80
---
kind: Service
apiVersion: v1
metadata:
  name: webstore-svc
spec:
  selector:
    app: webstore
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
`

type FieldType int

const (
	FieldUnknownType FieldType = iota
	FieldString
	FieldInt
	FieldBool
)

func (f *FieldType) UnmarshalMarkerArg(in string) error {
	types := map[string]FieldType{
		"":       FieldUnknownType,
		"string": FieldString,
		"int":    FieldInt,
		"bool":   FieldBool,
	}

	if t, ok := types[in]; ok {
		if t == FieldUnknownType {
			return fmt.Errorf("unable to parse %s into FieldType", in)
		}

		*f = t

		return nil
	}

	return fmt.Errorf("unable to parse %s into FieldType", in)
}

func (f FieldType) String() string {
	types := map[FieldType]string{
		FieldUnknownType: "",
		FieldString:      "string",
		FieldInt:         "int",
		FieldBool:        "bool",
	}

	return types[f]
}

type FieldMarker struct {
	Name        string
	Type        FieldType
	Description *string
	Default     interface{} `marker:",optional"`
}

type CollectionFieldMarker FieldMarker

func (fm FieldMarker) String() string {
	return fmt.Sprintf("FieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		fm.Name,
		fm.Type,
		*fm.Description,
		fm.Default,
	)
}

func main() {
	inspector, err := InitializeInspector()
	if err != nil {
		log.Fatal(err)
	}

	node, results, err := inspector.InspectYAML([]byte(code), TransformYAML)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		fmt.Println(result.Result.Object)
	}

	out, _ := yaml.Marshal(node)

	fmt.Print(string(out))
}

func InitializeInspector() (*inspect.Inspector, error) {
	registry := marker.NewRegistry()

	fieldMarker, err := marker.Define("+operator-builder:field", FieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	collectionMarker, err := marker.Define("+operator-builder:collection:field", CollectionFieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	registry.Add(fieldMarker)
	registry.Add(collectionMarker)

	return inspect.NewInspector(registry), nil
}

func TransformYAML(results ...inspect.YAMLResult) error {
	var key *yaml.Node

	var value *yaml.Node

	for _, r := range results {
		if len(r.Nodes) > 1 {
			key = r.Nodes[0]
			value = r.Nodes[1]
		} else {
			key = r.Nodes[0]
			value = r.Nodes[0]
		}

		key.HeadComment = ""
		key.FootComment = ""
		value.LineComment = ""

		switch t := r.Object.(type) {
		case FieldMarker:
			if *t.Description != "" {
				key.HeadComment = "# " + *t.Description + ", controlled by " + t.Name
			}

			value.Tag = "!!var"
			value.Value = fmt.Sprintf("parent.Spec." + t.Name)

		case CollectionFieldMarker:
			if *t.Description != "" {
				key.HeadComment = "# " + *t.Description + ", controlled by " + t.Name
			}

			value.Tag = "!!var"
			value.Value = fmt.Sprintf("collection.Spec." + t.Name)
		}

	}

	return nil
}
