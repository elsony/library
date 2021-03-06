# Devfile Parser Library

## About

The Devfile Parser library is a Golang module that:
1. parses the devfile.yaml as specified by the [api](https://devfile.github.io/devfile/api-reference.html) & [schema](https://github.com/devfile/api/tree/master/schemas/latest).
2. writes to the devfile.yaml with the updated data.
3. generates Kubernetes objects for the various devfile resources.
4. defines util functions for the devfile.

The function documentation can be accessed via [pkg.go.dev](https://pkg.go.dev/github.com/devfile/library). 
1. To parse a devfile, visit pkg/devfile/parse.go 
    ```
    // Parses the devfile and validates the devfile data
    devfile, err := devfilePkg.ParseAndValidate(devfileLocation)

    // To get all the components from the devfile
    components, err := devfile.Data.GetComponents(DevfileOptions{})
    ```
2. To get the Kubernetes objects from the devfile, visit pkg/devfile/generator/generators.go
   ```
    // To get a slice of Kubernetes containers of type corev1.Container from the devfile component containers
    containers, err := generator.GetContainers(devfile)

    // To generate a Kubernetes deployment of type v1.Deployment
    deployParams := generator.DeploymentParams{
		TypeMeta:          generator.GetTypeMeta(deploymentKind, deploymentAPIVersion),
		ObjectMeta:        generator.GetObjectMeta(name, namespace, labels, annotations),
		InitContainers:    initContainers,
		Containers:        containers,
		Volumes:           volumes,
		PodSelectorLabels: labels,
	}
	deployment := generator.GetDeployment(deployParams)
   ```

## Usage

In the future, the following projects will be consuming this library as a Golang dependency

* [Workspace Operator](https://github.com/devfile/devworkspace-operator)
* [odo](https://github.com/openshift/odo)
* [OpenShift Console](https://github.com/openshift/console)

## Issues

Issues are tracked in the [devfile/api](https://github.com/devfile/api) repo with the label [area/library](https://github.com/devfile/api/issues?q=is%3Aopen+is%3Aissue+label%3Aarea%2Flibrary) 