package target

import (
	"testing"

	"github.com/rancher/wrangler/pkg/yaml"

	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
)

const bundleYaml = `namespace: default
helm:
  releaseName: labels
  values:
    clusterName: global.fleet.clusterLabels.name
    customStruct:
      - name: global.fleet.clusterLabels.name
        key1: value1
        key2: value2
      - element1: global.fleet.clusterLabels.envType
      - element2: global.fleet.clusterLabels.name
diff:
  comparePatches:
  - apiVersion: networking.k8s.io/v1
    kind: Ingress
    name: labels-fleetlabelsdemo
    namespace: default
    operations:
    - op: remove
      path: /spec/rules/0/host
`

func TestProcessLabelValues(t *testing.T) {

	bundle := &v1alpha1.BundleSpec{}

	clusterLabels := make(map[string]string)
	clusterLabels["name"] = "local"
	clusterLabels["envType"] = "dev"

	err := yaml.Unmarshal([]byte(bundleYaml), bundle)
	if err != nil {
		t.Fatalf("error during yaml parsing %v", err)
	}

	err = processLabelValues(bundle.Helm.Values.Data, clusterLabels)
	if err != nil {
		t.Fatalf("error during label processing %v", err)
	}

	clusterName, ok := bundle.Helm.Values.Data["clusterName"]
	if !ok {
		t.Fatal("key clusterName not found")
	}

	if clusterName != "local" {
		t.Fatal("unable to assert correct clusterName")
	}

	customStruct, ok := bundle.Helm.Values.Data["customStruct"].([]interface{})
	if !ok {
		t.Fatal("key customStruct not found")
	}

	firstMap, ok := customStruct[0].(map[string]interface{})
	if !ok {
		t.Fatal("unable to assert first element to map[string]interface{}")
	}

	firstElemVal, ok := firstMap["name"]
	if !ok {
		t.Fatal("unable to find key name in the first element of customStruct")
	}

	if firstElemVal.(string) != "local" {
		t.Fatal("label replacement not performed in first element")
	}

	secondElement, ok := customStruct[1].(map[string]interface{})
	if !ok {
		t.Fatal("unable to assert second element of customStruct to map[string]interface{}")
	}

	secondElemVal, ok := secondElement["element1"]
	if !ok {
		t.Fatal("unable to find key element1")
	}

	if secondElemVal.(string) != "dev" {
		t.Fatal("label replacement not performed in second element")
	}

	thirdElement, ok := customStruct[2].(map[string]interface{})
	if !ok {
		t.Fatal("unable to assert third element of customStruct to map[string]interface{}")
	}

	thirdElemVal, ok := thirdElement["element2"]
	if !ok {
		t.Fatal("unable to find key element2")
	}

	if thirdElemVal.(string) != "local" {
		t.Fatal("label replacement not performed in third element")
	}
}

const bundleYamlWithTemplate = `namespace: default
helm:
  releaseName: labels
  values:
    clusterName: "{{ .ClusterLabels.name }}"
    fromAnnotation: "{{ .ClusterAnnotations.testAnnotation }}"
    clusterNamespace: "{{ .ClusterNamespace }}"
    fleetClusterName: "{{ .ClusterName }}"
    customStruct:
      - name: "{{ .Values.topLevel }}"
        key1: value1
        key2: value2
      - element2: "{{ .Values.nested.secondTier.thirdTier }}"
      - "element3_{{ .ClusterLabels.envType }}": "{{ .ClusterLabels.name }}"
    funcs:
      upper: "{{ .Values.topLevel | upper }}_test"
      join: '{{ .Values.list | join "," }}'
diff:
  comparePatches:
  - apiVersion: networking.k8s.io/v1
    kind: Ingress
    name: labels-fleetlabelsdemo
    namespace: default
    operations:
    - op: remove
      path: /spec/rules/0/host
`

func TestProcessTemplateValues(t *testing.T) {

	templateContext := map[string]interface{}{
		"topLevel": "foo",
		"nested": map[string]interface{}{
			"secondTier": map[string]interface{}{
				"thirdTier": "bar",
			},
		},
		"list": []string{
			"alpha",
			"beta",
			"omega",
		},
	}

	clusterLabels := map[string]string{
		"name":    "local",
		"envType": "dev",
	}

	clusterAnnotations := map[string]string{
		"testAnnotation": "test",
	}

	values := map[string]interface{}{
		"ClusterNamespace":   "dev-clusters",
		"ClusterName":        "my-cluster",
		"ClusterLabels":      clusterLabels,
		"ClusterAnnotations": clusterAnnotations,
		"Values":             templateContext,
	}

	bundle := &v1alpha1.BundleSpec{}
	err := yaml.Unmarshal([]byte(bundleYamlWithTemplate), bundle)
	if err != nil {
		t.Fatalf("error during yaml parsing %v", err)
	}

	templatedValues, err := processTemplateValues(bundle.Helm.Values.Data, values)
	if err != nil {
		t.Fatalf("error during label processing %v", err)
	}

	clusterName, ok := templatedValues["clusterName"]
	if !ok {
		t.Fatal("key clusterName not found")
	}

	if clusterName != "local" {
		t.Fatal("unable to assert correct clusterName")
	}

	fromAnnotation, ok := templatedValues["fromAnnotation"]
	if !ok {
		t.Fatal("key fromAnnotation not found")
	}

	if fromAnnotation != "test" {
		t.Fatal("unable to assert correct value for fromAnnotation")
	}

	clusterNamespace, ok := templatedValues["clusterNamespace"]
	if !ok {
		t.Fatal("key clusterNamespace not found")
	}

	if clusterNamespace != "dev-clusters" {
		t.Fatal("unable to assert correct value for clusterNamespace")
	}

	fleetClusterName, ok := templatedValues["fleetClusterName"]
	if !ok {
		t.Fatal("key clusterName not found")
	}

	if fleetClusterName != "my-cluster" {
		t.Fatal("unable to assert correct value fleetClusterName")
	}

	customStruct, ok := templatedValues["customStruct"].([]interface{})
	if !ok {
		t.Fatal("key customStruct not found")
	}

	firstMap, ok := customStruct[0].(map[string]interface{})
	if !ok {
		t.Fatal("unable to assert first element to map[string]interface{}")
	}

	firstElemVal, ok := firstMap["name"]
	if !ok {
		t.Fatal("unable to find key name in the first element of customStruct")
	}

	if firstElemVal.(string) != "foo" {
		t.Fatal("label replacement not performed in first element")
	}

	secondElement, ok := customStruct[1].(map[string]interface{})
	if !ok {
		t.Fatal("unable to assert second element of customStruct to map[string]interface{}")
	}

	secondElemVal, ok := secondElement["element2"]
	if !ok {
		t.Fatal("unable to find key element2")
	}

	if secondElemVal.(string) != "bar" {
		t.Fatal("template replacement not performed in second element")
	}

	thirdElement, ok := customStruct[2].(map[string]interface{})
	if !ok {
		t.Fatal("unable to assert second element of customStruct to map[string]interface{}")
	}

	thirdElemVal, ok := thirdElement["element3_dev"]
	if !ok {
		t.Fatal("unable to find key element3_dev")
	}

	if thirdElemVal.(string) != "local" {
		t.Fatal("template replacement not performed in third element")
	}

	funcs, ok := templatedValues["funcs"].(map[string]interface{})
	if !ok {
		t.Fatal("key funcs not found")
	}

	upper, ok := funcs["upper"]
	if !ok {
		t.Fatal("key upper not found")
	}

	if upper.(string) != "FOO_test" {
		t.Fatal("upper func was not right")
	}

	join, ok := funcs["join"]
	if !ok {
		t.Fatal("key join not found")
	}

	if join.(string) != "alpha,beta,omega" {
		t.Fatal("join func was not right")
	}

}
