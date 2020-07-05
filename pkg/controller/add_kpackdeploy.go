package controller

import (
	"github.com/cpage-pivotal/kpack-deploy-operator/pkg/controller/kpackdeploy"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kpackdeploy.Add)
}
