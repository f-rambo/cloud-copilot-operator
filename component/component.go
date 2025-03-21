package component

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	ComponentDir = "component"

	NetworkPolicy = filepath.Join(ComponentDir, "cilium-net-policy.yaml")
	Service       = filepath.Join(ComponentDir, "service.yaml")
	ConfigMap     = filepath.Join(ComponentDir, "configmap.yaml")
	Deployment    = filepath.Join(ComponentDir, "deployment.yaml")
	StatefulSet   = filepath.Join(ComponentDir, "statefulset.yaml")
	HttpRoute     = filepath.Join(ComponentDir, "http-route.yaml")

	Manifest = []string{
		ConfigMap,
		Deployment,
		StatefulSet,
		Service,
		HttpRoute,
		NetworkPolicy,
	}
)

func P() {
	fmt.Println(ComponentDir)
}

func GetManifest(data any) (*unstructured.UnstructuredList, error) {
	list := &unstructured.UnstructuredList{Items: make([]unstructured.Unstructured, 0)}
	for _, manifest := range Manifest {
		tempFile, err := transferredMeaning(data, manifest)
		if err != nil {
			return nil, err
		}
		if tempFile == "" {
			continue
		}
		unstructuredList, err := parseYAML(tempFile)
		if err != nil {
			return nil, err
		}
		for _, item := range unstructuredList.Items {
			if item.GetName() == "" || item.GetKind() == "" {
				continue
			}
			list.Items = append(list.Items, item)
		}
	}
	return list, nil
}

func transferredMeaning(data any, fileDetailPath string) (tmpFile string, err error) {
	if fileDetailPath == "" {
		return tmpFile, errors.New("fileDetailPath cannot be empty")
	}
	templateByte, err := os.ReadFile(fileDetailPath)
	if err != nil {
		return
	}
	tmpl, err := template.New(filepath.Base(fileDetailPath)).Funcs(template.FuncMap{
		"minus": func(a, b int32) int32 {
			return a - b
		},
		"return": func() string {
			return ""
		},
	}).Parse(string(templateByte))
	if err != nil {
		return
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return
	}

	tempFileName := "*-" + filepath.Base(fileDetailPath)
	tempDir := "/tmp"
	tmpFileObj, err := os.CreateTemp(tempDir, tempFileName)
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}
	defer tmpFileObj.Close()

	if err = os.Chmod(tmpFileObj.Name(), 0666); err != nil {
		return "", errors.Wrap(err, "failed to change file permissions")
	}

	if _, err = tmpFileObj.Write(buf.Bytes()); err != nil {
		return "", errors.Wrap(err, "failed to write to temp file")
	}

	return tmpFileObj.Name(), nil
}

func parseYAML(filename string) (*unstructured.UnstructuredList, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	list := &unstructured.UnstructuredList{Items: make([]unstructured.Unstructured, 0)}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	for {
		var obj map[string]any
		err = decoder.Decode(&obj)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		list.Items = append(list.Items, unstructured.Unstructured{Object: obj})
	}
	return list, nil
}
