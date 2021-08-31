# Markers

Operator Builder uses commented markers as the basis for defining a new API.
The fields for a custom resource kind are created when it finds a `+operator-builder`
marker in a source manifest.

A workload marker is commented out so the manifest is still valid and can be
used if needed.  The marker must begin with `+operator-builder` followed by some
colon-separated fields:

These markers should always be provided as an in-line comment or as a head comment.  The marker always begins with `+operator-builder:field:` or `+operator-builder:collection:field:` (more on this later).

That is followed by arguments separated by `,`.  Arguments can be given in any order.

## Field Marker
defined as `+operator-builder:field` this marker can be used to define a CRD field for your workload.

### Arguments 

arguments are separated from the marker name with a `:` they are given in the format of `argument=value` and separated by the `,`. additionally if the argument name is given by itself with no value, it is assumed to have an implict =true on the end and is treated as a flag.

below you will field the arguments for a field marker

#### Name (required)
the name you want to use for the field in the custom resource that
Operator Builder will create.  If you're not sure what that means, it will
become clear shortly.

ex. +operator-builder:field:name=myName

#### Type (required)
The other required field is the `type` field which specifies the data type for
the value.  The supported data types are:

- bool
- string
- int
- int32
- int64
- float32
- float64

ex. `+operator-builder:field:name=myName,type=string`

#### Default (optional)
This will make configuration optional for
your operator's end user. the supplied value will be used for the default value. If a field has no default, it will be a required field in the custom resource.

ex. `operator-builder:field:name=myName,type=string,default=test`
#### Description (optional)
An optional description can be provided which will be used in the source code as a Doc String, backticks `` ` `` may be used to capture multiline strings (head comments only)

By injecting documentation to
the CRD, the consumer of the custom resource gets the added benefit by being
able to run `kubectl explain` against their resource and having documentation
right at their fingertips without having to navigate to API documentation in
order to see the usage of the API.

ex. `operator-builder:field:name=myName,type=string,default=test,description="Hello World"`

Note that you can use a single custom resource field name to configure multiple
fields in the resource.  In the example above, the value for the `teamName`
field will.

Consider the following Deployment:

    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: webapp-deploy
      labels:
        production: false  # +operator-builder:field:name=production,default=false,type=bool
    spec:
      replicas: 2  # +operator-builder:field:name=webAppReplicas,default=2,type=int
      selector:
        matchLabels:
          app: webapp
      template:
        metadata:
          labels:
            app: webapp
        spec:
          containers:
          - name: webapp-container
            image: nginx:1.17  # +operator-builder:field:name=webAppImage,type=string
            ports:
            - containerPort: 8080

In this case, operator-builder will create and add three fields to the custom
resource:

- A `production` field that is a boolean.  It will have a default of `false` and
  will inform the value of the label when the deployment is configured.
- A `webAppReplicas` field that will default to `2` and allow the user to
  specify the number of replicas for the deployment in the custom resource
  manifest.
- A `webAppImage` field that will set the value for the images used in the pods.

Now the end-user of the operator will be able to define a custom resource
similar to the following to configure the deployment created:

	apiVersion: product.apps.acme.com/v1alpha1
	kind: WebApp
	metadata:
	  name: dev-webapp
	spec:
      production: false
      webAppReplicas: 2
      webAppImage: acmerepo/webapp:3.5.3


## Collection Markers

a second marker type `+operator-builder:collection:field` can be used with the same arguments as a Field Marker. These markers are used to define global Fields for your Collection and can be used in any of its associated components.
