//
// Copyright © 2021 Kris Nóva <kris@nivenly.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//   ███╗   ██╗ █████╗ ███╗   ███╗██╗
//   ████╗  ██║██╔══██╗████╗ ████║██║
//   ██╔██╗ ██║███████║██╔████╔██║██║
//   ██║╚██╗██║██╔══██║██║╚██╔╝██║██║
//   ██║ ╚████║██║  ██║██║ ╚═╝ ██║███████╗
//   ╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝
//

package codify

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/kris-nova/logger"
	rbacv1 "k8s.io/api/rbac/v1"
)

type RoleBinding struct {
	KubeObject *rbacv1.RoleBinding
	GoName     string
}

func NewRoleBinding(obj *rbacv1.RoleBinding) *RoleBinding {
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	return &RoleBinding{
		KubeObject: obj,
		GoName:     goName(obj.Name),
	}
}

func (k RoleBinding) Install() (string, []string) {
	c, err := Literal(k.KubeObject)
	if err != nil {
		logger.Debug(err.Error())
	}
	l := c.Source
	packages := c.Packages
	install := fmt.Sprintf(`
	{{ .GoName }}RoleBinding := %s
	x.objects = append(x.objects, {{ .GoName }}RoleBinding)

	if client != nil {
		_, err = client.RbacV1().RoleBindings("{{ .KubeObject.Namespace }}").Create(context.TODO(), {{ .GoName }}RoleBinding, v1.CreateOptions{})
		if err != nil {
			return err
		}
	}
`, l)

	tpl := template.New(fmt.Sprintf("%s", time.Now().String()))
	tpl.Parse(install)
	buf := &bytes.Buffer{}
	k.KubeObject.Name = sanitizeK8sObjectName(k.KubeObject.Name)
	err = tpl.Execute(buf, k)
	if err != nil {
		logger.Debug(err.Error())
	}
	return alias(buf.String(), "rbacv1"), packages
}

func (k RoleBinding) Uninstall() string {
	uninstall := `
	if client != nil {
		err = client.RbacV1().RoleBindings("{{ .KubeObject.Namespace }}").Delete(context.TODO(), "{{ .KubeObject.Name }}", metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
 `
	tpl := template.New(fmt.Sprintf("%s", time.Now().String()))
	tpl.Parse(uninstall)
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, k)
	if err != nil {
		logger.Debug(err.Error())
	}
	return buf.String()
}
