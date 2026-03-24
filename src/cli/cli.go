package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Context holds shared state for all CLI commands.
type Context struct {
	Lib        *protocol.Library
	InputFile  string
	OutputFile string
	Lang       string
	Raw        bool
}

// Run dispatches to the appropriate command handler.
func Run(ctx *Context, cmd string, args []string) error {
	switch cmd {
	case "list":
		return runList(ctx)
	case "show":
		if len(args) < 1 {
			return fmt.Errorf("usage: psl show <protocol>")
		}
		return runShow(ctx, args[0], ctx.Raw)
	case "decode":
		if len(args) < 1 {
			return fmt.Errorf("usage: psl decode <protocol>")
		}
		return runDecode(ctx, args[0])
	case "encode":
		if len(args) < 1 {
			return fmt.Errorf("usage: psl encode <protocol> [json]")
		}
		var jsonInput string
		if len(args) >= 2 {
			jsonInput = args[1]
		}
		return runEncode(ctx, args[0], jsonInput)
	case "validate":
		if len(args) < 1 {
			return fmt.Errorf("usage: psl validate <file.psl>")
		}
		return runValidate(ctx, args[0])
	case "pcap":
		return runPcap(ctx, args)
	case "generate":
		return runGenerate(ctx, args)
	case "diagram":
		return runDiagram(ctx, args)
	case "doc":
		return runDoc(ctx, args)
	case "fuzz":
		return runFuzz(ctx, args)
	case "install":
		return runInstall(ctx, args)
	case "uninstall":
		return runUninstall(ctx, args)
	case "plugins":
		return runPluginList(ctx)
	case "init":
		return runInitPackage(ctx, args)
	case "diff":
		return runDiff(ctx, args)
	case "compat":
		return runCompat(ctx, args)
	case "test":
		return runTest(ctx, args)
	case "audit":
		return runAudit(ctx, args)
	case "serve":
		return runServe(ctx, args)
	case "bench":
		return runBench(ctx, args)
	case "sdk":
		return runSDK(ctx, args)
	case "docsite":
		return runDocsite(ctx, args)
	case "convert":
		return runConvert(ctx, args)
	case "import":
		return runImport(ctx, args)
	case "mock":
		return runMock(ctx, args)
	case "loadtest":
		return runLoadTest(ctx, args)
	case "compliance":
		return runCompliance(ctx, args)
	case "monitor":
		return runMonitor(ctx, args)
	case "infer":
		return runInfer(ctx, args)
	case "playground":
		return runPlayground(ctx, args)
	case "identify":
		return runIdentify(ctx, args)
	case "project":
		return runProjectInit(ctx, args)
	case "githook":
		return runGitHook(ctx, args)
	case "sign":
		return runSign(ctx, args)
	case "pipe":
		return runPipe(ctx, args)
	case "diff-pcap":
		return runDiffPcap(ctx, args)
	case "sequence":
		return runSequence(ctx, args)
	case "apidoc":
		return runAPIDoc(ctx, args)
	case "testgen":
		return runTestGen(ctx, args)
	case "forensics":
		return runForensics(ctx, args)
	case "benchsuite":
		return runBenchSuite(ctx, args)
	case "cloud":
		return runCloud(ctx, args)
	case "decapsulate":
		return runDecapsulate(ctx, args)
	case "optimize":
		return runOptimize(ctx, args)
	case "ci":
		return runCI(ctx, args)
	case "pkg":
		return runPkg(ctx, args)
	case "migrate":
		return runMigrate(ctx, args)
	case "events":
		return runEvents(ctx, args)
	case "metrics":
		return runMetrics(ctx, args)
	case "etl":
		return runETL(ctx, args)
	case "constraint":
		return runConstraint(ctx, args)
	case "fsm":
		return runFSM(ctx, args)
	case "vdiff":
		return runVisualDiff(ctx, args)
	case "lint":
		return runLint(ctx, args)
	case "fmt":
		return runFmt(ctx, args)
	case "coverage":
		return runCoverage(ctx, args)
	case "repl":
		return runREPL(ctx, args)
	case "gateway":
		return runGateway(ctx, args)
	case "datalake":
		return runDatalake(ctx, args)
	case "notebook":
		return runNotebook(ctx, args)
	case "abtest":
		return runABTest(ctx, args)
	case "contract":
		return runContract(ctx, args)
	case "chaos":
		return runChaos(ctx, args)
	case "search":
		return runSearch(ctx, args)
	case "knowledge":
		return runKnowledge(ctx, args)
	case "tenant":
		return runTenant(ctx, args)
	case "standards":
		return runStandards(ctx, args)
	case "edu":
		return runEdu(ctx, args)
	case "ecosystem":
		return runEcosystem(ctx, args)
	case "ebpf":
		return runEBPF(ctx, args)
	case "dpdk":
		return runDPDK(ctx, args)
	case "p4":
		return runP4Gen(ctx, args)
	case "mesh":
		return runMesh(ctx, args)
	case "pqcrypto":
		return runPQCrypto(ctx, args)
	case "twin":
		return runTwin(ctx, args)
	case "ai":
		return runAIGen(ctx, args)
	case "autocomply":
		return runAutoComply(ctx, args)
	case "desktop":
		return runDesktop(ctx, args)
	case "mobile":
		return runMobile(ctx, args)
	case "psl4":
		return runPSL4(ctx, args)
	default:
		return fmt.Errorf("unknown command: %s\nRun 'psl -h' for usage", cmd)
	}
}
