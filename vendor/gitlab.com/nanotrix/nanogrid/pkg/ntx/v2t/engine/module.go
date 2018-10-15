package engine

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
)

const LABEL_SEPARATOR = "+"

func ModulesToLabel(modules ...EngineModule) string {
	res := []string{}

	modules = SortModules(DeduplicateModules(modules))

	for _, m := range modules {
		res = append(res, m.ShortName())
	}

	return strings.Join(res, LABEL_SEPARATOR)
}

func ContainsModule(mod EngineModule, mods []EngineModule) bool {
	for _, l := range mods {
		if l == mod {
			return true
		}
	}
	return false
}

func (label EngineModule) ShortName() string {
	return strings.ToLower(strings.Replace(label.String(), "MODULE_", "", 1))
}

func EngineModuleForShortName(simpleName string) (EngineModule, bool) {
	module, found := EngineModule_value["MODULE_"+strings.ToUpper(simpleName)]
	return EngineModule(module), found
}

func SortModules(mods []EngineModule) []EngineModule {
	sort.Slice(mods, func(i, j int) bool {
		return mods[i] < mods[j]
	})

	return mods
}

func DeduplicateModules(mods []EngineModule) []EngineModule {
	dup := make(map[EngineModule]bool)

	for _, m := range mods {
		dup[m] = true
	}

	res := make([]EngineModule, len(dup))

	i := 0
	for k := range dup {
		res[i] = k
		i++
	}

	return res
}

func ModulesFromLabel(label string) ([]EngineModule, error) {
	if label == "" {
		return nil, nil
	}

	modsS := strings.Split(strings.ToLower(label), LABEL_SEPARATOR)
	modules := make([]EngineModule, len(modsS))

	for i, s := range modsS {
		m, ok := EngineModuleForShortName(s)
		if !ok {
			return nil, errors.Errorf("Invalid module short name '%v' in '%v'", s, label)
		}
		modules[i] = m
	}

	return SortModules(DeduplicateModules(modules)), nil
}

func ModuleToEngineContext(mods []EngineModule) (*EngineContext, error) {
	if len(mods) == 0 {
		return nil, errors.New("Cannot create context from empty modules")
	}

	vad := ContainsModule(EngineModule_MODULE_VAD, mods)
	v2t := ContainsModule(EngineModule_MODULE_V2T, mods)
	ppc := ContainsModule(EngineModule_MODULE_PPC, mods)
	pnc := ContainsModule(EngineModule_MODULE_PNC, mods)

	if vad && ppc && !v2t {
		return nil, errors.New("Invalid module selection vad+ppc")
	}

	if pnc && !v2t {
		return nil, errors.New("v2t required for pnc")
	}

	ctx := &EngineContext{
		AudioFormat: &AudioFormat{Formats: &AudioFormat_Auto{Auto: &AudioFormat_AutoDetect{}}},
	}

	if v2t {
		var vadCfg *EngineContext_VADConfig
		var ppcCfg *EngineContext_PPCConfig
		var pncCfg *EngineContext_PNCConfig
		if vad {
			vadCfg = &EngineContext_VADConfig{}
		}

		if ppc {
			ppcCfg = &EngineContext_PPCConfig{}
		}

		if pnc {
			pncCfg = &EngineContext_PNCConfig{}
		}

		ctx.Config = &EngineContext_V2T{
			V2T: &EngineContext_V2TConfig{
				WithVAD: vadCfg,
				WithPPC: ppcCfg,
				WithPNC: pncCfg,
			},
		}
	} else if vad {
		ctx.Config = &EngineContext_Vad{Vad: &EngineContext_VADConfig{}}
	} else if ppc {
		ctx.Config = &EngineContext_Ppc{Ppc: &EngineContext_PPCConfig{}}
	} else {
		return nil, errors.New("No known module selected")
	}

	return ctx, nil
}
