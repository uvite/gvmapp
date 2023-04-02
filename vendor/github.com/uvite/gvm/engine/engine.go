package engine

import (
	"context"
	"fmt"
	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/uvite/gvm/pkg/js"
	"github.com/uvite/gvm/pkg/lib"
	"github.com/uvite/gvm/pkg/lib/consts"
	"github.com/uvite/gvm/pkg/loader"
	"github.com/uvite/gvm/pkg/metrics"
	_ "github.com/uvite/gvm/plugin/xk6-dotenv"
	_ "github.com/uvite/gvm/plugin/xk6-exec"
	_ "github.com/uvite/gvm/plugin/xk6-file"
	_ "github.com/uvite/gvm/plugin/xk6-nats"
	_ "github.com/uvite/gvm/plugin/xk6-redis"
	//_ "github.com/uvite/gvm/plugin/xk6-sql"
	_ "github.com/uvite/gvm/plugin/xk6-ta"
	_ "github.com/uvite/gvm/plugin/xk6-timers"
	_ "github.com/uvite/gvm/plugin/xk6-websockets"
	_ "github.com/uvite/gvm/plugin/xk6-yaml"
	"os"
)

type Gvm struct {
	Ctx context.Context
	*goja.Runtime
	Vu        lib.ActiveVU
	Logger    *logrus.Logger
	Runner    *js.Runner
	CancelFun context.CancelFunc
}

func NewGvm() (*Gvm, error) {
	gvm := Gvm{}
	return &gvm, nil
}

func (gvm *Gvm) LoadFile(filepath string) error {

	fs := afero.NewOsFs()
	pwd, _ := os.Getwd()
	logger := logrus.New()

	gvm.Logger = logger
	//filepath := fmt.Sprintf("%s/%s", pwd, file)
	code, err := loader.ReadSource(logger, filepath, pwd, map[string]afero.Fs{"file": fs}, nil)
	if err != nil {
		return fmt.Errorf("couldn't load file: %s", err)
	}
	//fmt.Println(filepath)
	//fmt.Println(code.Data)

	rtOpts := lib.RuntimeOptions{}
	r, err := gvm.GetSimpleRunner(filepath, fmt.Sprintf(`
			import {Nats} from 'k6/x/nats';
			import ta from 'k6/x/ta';
			import {sleep} from 'k6'; 
			exports.options = {
				setupTimeout: "5s",
			};
		 

			%s

			`, code.Data),
		fs, rtOpts)

	//fmt.Println(err)

	gvm.Runner = r
	gvm.Runtime = r.Bundle.Vm
	if err != nil {
		return fmt.Errorf("couldn't set exported options with merged values: %w", err)

	}

	return nil
}

func (gvm *Gvm) Load(file string) error {

	fs := afero.NewOsFs()
	pwd, _ := os.Getwd()
	logger := logrus.New()

	gvm.Logger = logger
	filepath := fmt.Sprintf("%s/%s", pwd, file)
	code, err := loader.ReadSource(logger, filepath, pwd, map[string]afero.Fs{"file": fs}, nil)
	if err != nil {
		return fmt.Errorf("couldn't load file: %s", err)
	}
	fmt.Println(filepath)
	fmt.Println(code.Data)

	rtOpts := lib.RuntimeOptions{}
	r, err := gvm.GetSimpleRunner(filepath, fmt.Sprintf(`
			import {Nats} from 'k6/x/nats';
			import ta from 'k6/x/ta';
			import {sleep} from 'k6'; 
			exports.options = {
				setupTimeout: "5s",
			}; 
			%s

			`, code.Data),
		fs, rtOpts)

	//fmt.Println(err)

	gvm.Runner = r
	gvm.Runtime = r.Bundle.Vm
	if err != nil {
		return fmt.Errorf("couldn't set exported options with merged values: %w", err)

	}

	return nil
}

func (gvm *Gvm) LoadFiel(filepath string) error {

	fs := afero.NewOsFs()
	pwd, _ := os.Getwd()
	logger := logrus.New()

	gvm.Logger = logger
	//filepath := fmt.Sprintf("%s/%s", pwd, file)
	code, err := loader.ReadSource(logger, filepath, pwd, map[string]afero.Fs{"file": fs}, nil)
	if err != nil {
		return fmt.Errorf("couldn't load file: %s", err)
	}
	fmt.Println(filepath)
	fmt.Println(code.Data)

	rtOpts := lib.RuntimeOptions{}
	r, err := gvm.GetSimpleRunner(filepath, fmt.Sprintf(`
			import {Nats} from 'k6/x/nats';
			import ta from 'k6/x/ta';
			import {sleep} from 'k6'; 
			exports.options = {
				setupTimeout: "5s",
			}; 
			%s

			`, code.Data),
		fs, rtOpts)

	//fmt.Println(err)

	gvm.Runner = r
	gvm.Runtime = r.Bundle.Vm
	if err != nil {
		return fmt.Errorf("couldn't set exported options with merged values: %w", err)

	}

	return nil
}

// func (gvm *Gvm) Init() error {
//
//		ch := make(chan metrics.SampleContainer, 100)
//		//ctx, _ := context.WithCancel(context.Background())
//		//defer cancel()
//		err := gvm.Runner.Setup(gvm.Ctx, ch)
//		if err != nil {
//			return err
//		}
//		initVU, err := gvm.Runner.NewVU(gvm.Ctx, 1, 1, ch)
//
//		vu := initVU.Activate(&lib.VUActivationParams{RunContext: gvm.Ctx})
//		gvm.Vu = vu
//		return err
//	}
func (gvm *Gvm) Init() (context.CancelFunc, error) {

	ch := make(chan metrics.SampleContainer, 100)
	ctx, cancel := context.WithCancel(gvm.Ctx)
	//defer cancel()
	gvm.CancelFun = cancel

	err := gvm.Runner.Setup(ctx, ch)
	fmt.Println(err)
	if err != nil {
		return cancel, err
	}
	initVU, err := gvm.Runner.NewVU(ctx, 1, 1, ch)

	vu := initVU.Activate(&lib.VUActivationParams{RunContext: ctx})
	gvm.Vu = vu
	return cancel, err
}

func (gvm *Gvm) Run() (goja.Value, error) {

	//gvm.Vu.RunOnce()
	v, ok := gvm.Vu.RunDefault()
	//fmt.Println(v, ok)
	if ok != nil {
		return nil, ok
	}
	return v, nil
}
func (gvm *Gvm) Set(name string, value any) {
	err := gvm.Runner.Bundle.Vm.Set(name, value)
	if err != nil {
		fmt.Println(err)
	}

}

func (gvm *Gvm) ExecFunc(fun string) (goja.Value, error) {
	r := gvm.Runner
	if !r.IsExecutable(fun) {
		// do not init a new transient VU or execute setup() if it wasn't
		// actually defined and exported in the script
		gvm.Logger.Debugf("%s() is not defined or not exported, skipping!", fun)
		return nil, nil
	}
	gvm.Logger.Debugf("Running %s()...", consts.SetupFn)
	ctx, cancel := context.WithCancel(gvm.Ctx)
	defer cancel()

	out := make(chan metrics.SampleContainer, 100)
	v, err := r.RunPart(ctx, out, fun, nil)
	if err != nil {
		return nil, err
	}
	// r.setupData = nil is special it means undefined from this moment forward
	if goja.IsUndefined(v) {

		return nil, nil
	}
	return v, err
}
