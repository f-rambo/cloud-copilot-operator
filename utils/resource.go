package utils

import (
	"bytes"
	"os"
	"text/template"

	operatoroceaniov1alpha1 "github.com/f-rambo/operatorapp/api/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// 判断文件是否存在
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// 创建一个可读可写的文件
func CreateFile(filename string) error {
	var file, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

// 创建一个嵌套的目录
func CreateDir(path string) error {
	var err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func parseTemplate(templateName string, app *operatoroceaniov1alpha1.App) ([]byte, error) {
	tmpl, err := template.ParseFiles("internal/template/" + templateName + ".yml")
	if err != nil {
		return nil, err
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, app)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func NewDeployment(app *operatoroceaniov1alpha1.App) (*appv1.Deployment, error) {
	d := &appv1.Deployment{}
	data, err := parseTemplate("deployment", app)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func NewIngress(app *operatoroceaniov1alpha1.App) (*netv1.Ingress, error) {
	i := &netv1.Ingress{}
	data, err := parseTemplate("ingress", app)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func NewService(app *operatoroceaniov1alpha1.App) (*corev1.Service, error) {
	s := &corev1.Service{}
	data, err := parseTemplate("service", app)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
