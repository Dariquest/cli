package push

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a simple manifest", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	When("the manifest is in the current directory", func() {
		It("uses the manifest", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"env": map[string]interface{}{
								"key1": "val1",
								"key4": false,
							},
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("env", appName)
			Eventually(session).Should(Say(`key1:\s+val1`))
			Eventually(session).Should(Say(`key4:\s+false`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the manifest is provided via '-f' flag", func() {
		var (
			tempDir        string
			pathToManifest string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "simple-manifest-test")
			Expect(err).ToNot(HaveOccurred())
			pathToManifest = filepath.Join(tempDir, "manifest.yml")
			helpers.WriteManifest(pathToManifest, map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"env": map[string]interface{}{
							"key1": "val1",
							"key4": false,
						},
					},
				},
			})
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
		})

		It("uses the manifest", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(
					helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-f", pathToManifest,
					"--no-start")
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("env", appName)
			Eventually(session).Should(Say(`key1:\s+val1`))
			Eventually(session).Should(Say(`key4:\s+false`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("A single --vars-file is specified", func() {
		var (
			tempDir        string
			pathToManifest string
			pathToVarsFile string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "simple-manifest-test")
			Expect(err).ToNot(HaveOccurred())
			pathToManifest = filepath.Join(tempDir, "manifest.yml")
			helpers.WriteManifest(pathToManifest, map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"env": map[string]interface{}{
							"key1": "((var1))",
							"key4": false,
						},
					},
				},
			})

			pathToVarsFile = filepath.Join(tempDir, "vars.yml")
			helpers.WriteManifest(pathToVarsFile, map[string]interface{}{"var1": "secret-key"})
		})

		It("uses the manifest with substituted variables", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(
					helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-f", pathToManifest,
					"--vars-file", pathToVarsFile,
					"--no-start")
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("env", appName)
			Eventually(session).Should(Say(`key1:\s+secret-key`))
			Eventually(session).Should(Say(`key4:\s+false`))
			Eventually(session).Should(Exit(0))
		})

	})

	When("A multiple --vars-file are specified", func() {
		var (
			tempDir         string
			pathToManifest  string
			pathToVarsFile1 string
			pathToVarsFile2 string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "simple-manifest-test")
			Expect(err).ToNot(HaveOccurred())
			pathToManifest = filepath.Join(tempDir, "manifest.yml")
			helpers.WriteManifest(pathToManifest, map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"env": map[string]interface{}{
							"key1": "((var1))",
							"key4": "((var2))",
						},
					},
				},
			})

			pathToVarsFile1 = filepath.Join(tempDir, "vars1.yml")
			helpers.WriteManifest(pathToVarsFile1, map[string]interface{}{"var1": "not-so-secret-key", "var2": "foobar"})

			pathToVarsFile2 = filepath.Join(tempDir, "vars2.yml")
			helpers.WriteManifest(pathToVarsFile2, map[string]interface{}{"var1": "secret-key"})
		})

		It("uses the manifest with substituted variables", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(
					helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-f", pathToManifest,
					"--vars-file", pathToVarsFile1,
					"--vars-file", pathToVarsFile2,
					"--no-start")
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("env", appName)
			Eventually(session).Should(Say(`key1:\s+secret-key`))
			Eventually(session).Should(Say(`key4:\s+foobar`))
			Eventually(session).Should(Exit(0))
		})

	})

	When("A manifest contains a variable not in the vars-file", func() {
		var (
			tempDir         string
			pathToManifest  string
			pathToVarsFile1 string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "simple-manifest-test")
			Expect(err).ToNot(HaveOccurred())
			pathToManifest = filepath.Join(tempDir, "manifest.yml")
			helpers.WriteManifest(pathToManifest, map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"env": map[string]interface{}{
							"key1": "((var1))",
							"key4": "((var2))",
						},
					},
				},
			})

			pathToVarsFile1 = filepath.Join(tempDir, "vars1.yml")
			helpers.WriteManifest(pathToVarsFile1, map[string]interface{}{"var2": "foobar"})
		})

		It("It fails to push because of unfound variables", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(
					helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-f", pathToManifest,
					"--vars-file", pathToVarsFile1,
					"--no-start")
				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say(`Expected to find variables: var1`))
			})
		})

	})
})
