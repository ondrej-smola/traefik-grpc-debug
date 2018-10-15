package license

const (
	EngineLicenseEnvKey = "ntx4_licence.audience"
)

func UpdateTaskEnvFnProvider(gen EngineLicenseGenerateFn) func(map[string]string) {
	return func(env map[string]string) {
		env[EngineLicenseEnvKey] = gen()
	}
}
