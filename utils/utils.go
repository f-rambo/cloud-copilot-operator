package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func FindMatchingFile(dir, pattern string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"+pattern+"*"))
	if err != nil {
		return "", errors.Wrap(err, "failed to search files")
	}
	for _, f := range files {
		if !strings.HasSuffix(f, ".tmp") {
			return f, nil
		}
	}
	return "", nil
}

var (
	NetworkPolicy = "cilium-net-policy.yaml"
	Service       = "service.yaml"
	ConfigMap     = "configmap.yaml"
	Deployment    = "deployment.yaml"
	StatefulSet   = "statefulset.yaml"
	HttpRoute     = "http-route.yaml"
	GrpcRoute     = "grpc-route.yaml"

	CloudServiceManifest = []string{
		ConfigMap,
		Deployment,
		StatefulSet,
		Service,
		HttpRoute,
		GrpcRoute,
		NetworkPolicy,
	}
)

func GetCloudServiceManifest(manifestdir string, data any) (*unstructured.UnstructuredList, error) {
	list := &unstructured.UnstructuredList{Items: make([]unstructured.Unstructured, 0)}
	for _, manifest := range CloudServiceManifest {
		if manifestdir != "" {
			manifest = filepath.Join(manifestdir, manifest)
		}
		tempFile, err := TransferredMeaning(data, manifest)
		if err != nil {
			return nil, err
		}
		if tempFile == "" {
			continue
		}
		unstructuredList, err := ParseYAML(tempFile)
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

func TransferredMeaning(data any, fileDetailPath string) (tmpFile string, err error) {
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

	tmpFileObj, err := os.CreateTemp("/tmp", "*-"+filepath.Base(fileDetailPath))
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

func ParseYAML(filename string) (*unstructured.UnstructuredList, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewYAMLOrJSONDecoder(file, 4096)
	list := &unstructured.UnstructuredList{Items: make([]unstructured.Unstructured, 0)}
	for {
		var obj map[string]any
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		if len(obj) == 0 {
			continue
		}
		list.Items = append(list.Items, unstructured.Unstructured{Object: obj})
	}
	return list, nil
}

func GetKubeClientByKubeConfig(KubeConfigPaths ...string) (clientset *kubernetes.Clientset, err error) {
	var KubeConfigPath string
	if len(KubeConfigPaths) == 0 {
		KubeConfigPath = clientcmd.RecommendedHomeFile
	} else {
		KubeConfigPath = KubeConfigPaths[0]
	}
	config, err := clientcmd.BuildConfigFromFlags("", KubeConfigPath)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "get kubernetes by kubeconfig client failed")
	}
	return client, nil
}

// labels (string) to map[string]string
func LabelsToMap(labels string) map[string]string {
	if labels == "" {
		return make(map[string]string)
	}
	m := make(map[string]string)
	for _, label := range strings.Split(labels, ",") {
		kv := strings.Split(label, "=")
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return m
}

// map[string]string to labels (string)
func MapToLabels(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	var labels []string
	for k, v := range m {
		labels = append(labels, k+"="+v)
	}
	return strings.Join(labels, ",")
}

func PointerBool(b bool) *bool {
	return &b
}
