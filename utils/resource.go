package utils

import (
	"bytes"
	"os"
	"text/template"

	operatorv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	SecretKey = "secret"
	ConfigKey = "config"
)

// 判断文件是否存在
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func CreateFile(filename string) error {
	var file, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func CreateDir(path string) error {
	var err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func parseTemplate(templateName string, cloudService *operatorv1alpha1.CloudService) ([]byte, error) {
	tmpl, err := template.ParseFiles("internal/template/" + templateName + ".yml")
	if err != nil {
		return nil, err
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, cloudService)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func NewDeployment(cloudService *operatorv1alpha1.CloudService) (*appv1.Deployment, error) {
	d := &appv1.Deployment{}
	data, err := parseTemplate("deployment", cloudService)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func NewIngress(cloudService *operatorv1alpha1.CloudService) (*netv1.Ingress, error) {
	i := &netv1.Ingress{}
	data, err := parseTemplate("ingress", cloudService)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func NewService(cloudService *operatorv1alpha1.CloudService) (*corev1.Service, error) {
	s := &corev1.Service{}
	data, err := parseTemplate("service", cloudService)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
