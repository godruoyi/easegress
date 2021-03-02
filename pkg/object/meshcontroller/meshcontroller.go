package meshcontroller

import (
	"github.com/megaease/easegateway/pkg/logger"
	"github.com/megaease/easegateway/pkg/option"
	"github.com/megaease/easegateway/pkg/supervisor"
)

const (
	// Category is the category of MeshController.
	Category = supervisor.CategoryBusinessController

	// Kind is the kind of MeshController.
	Kind = "MeshController"
)

type (
	// MeshController is a business controller to complete MegaEase Service Mesh.
	MeshController struct {
		super     *supervisor.Supervisor
		superSpec *supervisor.Spec
		spec      *Spec

		role   string
		master *Master
		worker *Worker
	}
)

func init() {
	supervisor.Register(&MeshController{})
}

// Category returns the category of MeshController.
func (mc *MeshController) Category() supervisor.ObjectCategory {
	return Category
}

// Kind return the kind of MeshController.
func (mc *MeshController) Kind() string {
	return Kind
}

// DefaultSpec returns the default spec of MeshController.
func (mc *MeshController) DefaultSpec() interface{} {
	return &Spec{
		SpecUpdateInterval: "10s",
		HeartbeatInterval:  "5s",
		RegistryType:       "consul",
	}

}

// Init initializes MeshController.
func (mc *MeshController) Init(superSpec *supervisor.Spec, super *supervisor.Supervisor) {
	mc.superSpec, mc.spec, mc.super = superSpec, superSpec.ObjectSpec().(*Spec), super
	mc.reload()
}

// Inherit inherits previous generation of MeshController.
func (mc *MeshController) Inherit(spec *supervisor.Spec,
	previousGeneration supervisor.Object, super *supervisor.Supervisor) {

	previousGeneration.Close()
	mc.Init(spec, super)
}

func (mc *MeshController) reload() {
	role := option.Global.Labels["mesh_role"]
	switch role {
	case meshRoleMaster:
		logger.Infof("%s running in master role", mc.superSpec.Name())
		mc.role = meshRoleMaster
	case meshRoleWorker:
		logger.Infof("%s running in worker role", mc.superSpec.Name())
		mc.role = meshRoleWorker
	default:
		logger.Infof("%s running in master role (default mode)", mc.superSpec.Name())
		mc.role = meshRoleMaster
	}

	if mc.role == meshRoleMaster {
		mc.master = NewMaster(mc.superSpec, mc.super)
		return
	}

	mc.worker = NewWorker(mc.superSpec, mc.super)
}

// Status returns the status of MeshController.
func (mc *MeshController) Status() *supervisor.Status {
	if mc.master != nil {
		return mc.master.Status()
	}

	return mc.worker.Status()
}

// Close closes MeshController.
func (mc *MeshController) Close() {
	if mc.master != nil {
		mc.master.Close()
		return
	}

	mc.worker.Close()
}
