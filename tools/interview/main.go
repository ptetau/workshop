package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	defaultOutputBase = ".scaffold/interview/graph"
)

type field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type concept struct {
	Name    string  `json:"name"`
	Fields  []field `json:"fields"`
	Methods []string
}

type orchestrator struct {
	Name   string  `json:"name"`
	Params []field `json:"params"`
	Route  *route  `json:"route,omitempty"`
}

type projection struct {
	Name   string  `json:"name"`
	Query  []field `json:"query"`
	Result []field `json:"result"`
	Route  *route  `json:"route,omitempty"`
}

type route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Target string `json:"target"`
}

type graph struct {
	GeneratedAt   time.Time      `json:"generatedAt"`
	Source        string         `json:"source"`
	Concepts      []concept      `json:"concepts"`
	Orchestrators []orchestrator `json:"orchestrators"`
	Projections   []projection   `json:"projections"`
	Routes        []route        `json:"routes"`
	ScaffoldArgs  []string       `json:"scaffoldArgs"`
}

func main() {
	var (
		root      string
		module    string
		outBase   string
		threshold float64
	)

	flag.StringVar(&root, "root", ".", "output directory for scaffold")
	flag.StringVar(&module, "module", "", "module path for scaffold")
	flag.StringVar(&outBase, "out", defaultOutputBase, "output base path for graph files")
	flag.Float64Var(&threshold, "threshold", 0.8, "similarity threshold for disambiguation")
	flag.Parse()

	interviewer := newInterviewer(os.Stdin, os.Stdout, threshold)
	g, err := interviewer.run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "interview:", err)
		os.Exit(1)
	}

	args := buildScaffoldArgs(g)
	g.ScaffoldArgs = args

	if err := writeGraph(outBase, g); err != nil {
		fmt.Fprintln(os.Stderr, "interview:", err)
		os.Exit(1)
	}

	if err := runScaffold(root, module, args); err != nil {
		fmt.Fprintln(os.Stderr, "interview:", err)
		os.Exit(1)
	}
}

type interviewer struct {
	in        *bufio.Scanner
	out       *bufio.Writer
	threshold float64
}

func newInterviewer(in io.Reader, out io.Writer, threshold float64) *interviewer {
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)
	return &interviewer{
		in:        scanner,
		out:       bufio.NewWriter(out),
		threshold: threshold,
	}
}

func (i *interviewer) run() (graph, error) {
	g := graph{
		GeneratedAt: time.Now(),
		Source:      "Interactive",
	}

	i.say("Starting interview. Type 'done' to finish a section.")

	concepts := i.askConcepts()
	orchestrators := i.askOrchestrators()
	projections := i.askProjections()

	g.Concepts = concepts
	g.Orchestrators = orchestrators
	g.Projections = projections
	g.Routes = collectRoutes(orchestrators, projections)

	return g, i.out.Flush()
}

func (i *interviewer) askConcepts() []concept {
	var concepts []concept
	for {
		name, ok := i.askSingleWord("Concept (one word)", "")
		if !ok {
			break
		}
		symbol := symbolify(name)
		if existing, merged := i.disambiguate("concept", symbol, conceptNames(concepts)); merged {
			i.say("Using existing concept %s.", existing)
			continue
		}
		c := concept{Name: symbol}
		c.Fields = i.askFields(symbol, "Field")
		c.Methods = i.askMethods(symbol)
		concepts = append(concepts, c)
	}
	return concepts
}

func (i *interviewer) askOrchestrators() []orchestrator {
	var orchestrators []orchestrator
	for {
		name, ok := i.askPhrase("Orchestrator (short phrase)")
		if !ok {
			break
		}
		symbol := symbolify(name)
		if existing, merged := i.disambiguate("orchestrator", symbol, orchestratorNames(orchestrators)); merged {
			i.say("Using existing orchestrator %s.", existing)
			continue
		}
		o := orchestrator{Name: symbol}
		o.Params = i.askFields(symbol, "Param")
		o.Route = i.askCommandRoute(symbol)
		orchestrators = append(orchestrators, o)
	}
	return orchestrators
}

func (i *interviewer) askProjections() []projection {
	var projections []projection
	for {
		name, ok := i.askPhrase("Projection (short phrase)")
		if !ok {
			break
		}
		symbol := symbolify(name)
		if existing, merged := i.disambiguate("projection", symbol, projectionNames(projections)); merged {
			i.say("Using existing projection %s.", existing)
			continue
		}
		p := projection{Name: symbol}
		p.Query = i.askFields(symbol, "Query field")
		p.Result = i.askFields(symbol, "Result field")
		p.Route = i.askQueryRoute(symbol)
		projections = append(projections, p)
	}
	return projections
}

func (i *interviewer) askFields(owner, label string) []field {
	var fields []field
	for {
		name, ok := i.askSingleWord(fmt.Sprintf("%s for %s", label, owner), "")
		if !ok {
			break
		}
		typ, _ := i.askSingleWord("Type (string/int/bool/time/custom)", "string")
		fields = append(fields, field{Name: symbolify(name), Type: typ})
	}
	return fields
}

func (i *interviewer) askMethods(owner string) []string {
	var methods []string
	for {
		name, ok := i.askSingleWord(fmt.Sprintf("Method for %s", owner), "")
		if !ok {
			break
		}
		methods = append(methods, symbolify(name))
	}
	return methods
}

func (i *interviewer) askCommandRoute(owner string) *route {
	method, ok := i.askSingleWord("Route method for "+owner+" (POST/PUT/DELETE/skip)", "skip")
	if !ok || strings.EqualFold(method, "skip") {
		return nil
	}
	path := i.askLine("Route path for " + owner + " (e.g. /orders)")
	if path == "" {
		return nil
	}
	return &route{Method: strings.ToUpper(method), Path: path, Target: owner}
}

func (i *interviewer) askQueryRoute(owner string) *route {
	path := i.askLine("Route path for " + owner + " (GET) or 'skip'")
	if strings.EqualFold(strings.TrimSpace(path), "skip") || path == "" {
		return nil
	}
	return &route{Method: "GET", Path: path, Target: owner}
}

func (i *interviewer) disambiguate(kind, candidate string, existing []string) (string, bool) {
	for _, other := range existing {
		if similarity(candidate, other) >= i.threshold {
			yes, _ := i.askYesNo(fmt.Sprintf("Is %s %q the same as %q", kind, candidate, other))
			if yes {
				return other, true
			}
		}
	}
	return candidate, false
}

func (i *interviewer) askYesNo(prompt string) (bool, bool) {
	for {
		answer := strings.ToLower(i.askLine(prompt + " (yes/no)"))
		switch answer {
		case "yes":
			return true, true
		case "no":
			return false, true
		case "done":
			return false, false
		default:
			i.say("Please answer 'yes' or 'no'.")
		}
	}
}

func (i *interviewer) askSingleWord(prompt string, defaultValue string) (string, bool) {
	for {
		answer := strings.TrimSpace(i.askLine(prompt))
		if answer == "" && defaultValue != "" {
			return defaultValue, true
		}
		if strings.EqualFold(answer, "done") {
			return "", false
		}
		if strings.Contains(answer, " ") {
			i.say("One word only.")
			continue
		}
		return answer, true
	}
}

func (i *interviewer) askPhrase(prompt string) (string, bool) {
	for {
		answer := strings.TrimSpace(i.askLine(prompt))
		if strings.EqualFold(answer, "done") {
			return "", false
		}
		if answer == "" {
			i.say("Please provide a short phrase or 'done'.")
			continue
		}
		return answer, true
	}
}

func (i *interviewer) askLine(prompt string) string {
	i.say(prompt + ":")
	if !i.in.Scan() {
		return ""
	}
	return strings.TrimSpace(i.in.Text())
}

func (i *interviewer) say(format string, args ...any) {
	fmt.Fprintf(i.out, format+"\n", args...)
	i.out.Flush()
}

func writeGraph(outBase string, g graph) error {
	if outBase == "" {
		outBase = defaultOutputBase
	}
	if err := os.MkdirAll(filepath.Dir(outBase), 0o755); err != nil {
		return err
	}

	jsonPath := outBase + ".json"
	jsonData, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return err
	}

	dotPath := outBase + ".dot"
	if err := os.WriteFile(dotPath, []byte(buildDot(g)), 0o644); err != nil {
		return err
	}

	return nil
}

func buildDot(g graph) string {
	var b strings.Builder
	b.WriteString("digraph Scaffold {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  node [shape=box];\n")

	for _, c := range g.Concepts {
		b.WriteString(fmt.Sprintf("  \"%s\" [shape=box];\n", c.Name))
	}
	for _, o := range g.Orchestrators {
		b.WriteString(fmt.Sprintf("  \"%s\" [shape=ellipse];\n", o.Name))
	}
	for _, p := range g.Projections {
		b.WriteString(fmt.Sprintf("  \"%s\" [shape=diamond];\n", p.Name))
	}
	for _, r := range g.Routes {
		node := fmt.Sprintf("%s %s", r.Method, r.Path)
		b.WriteString(fmt.Sprintf("  \"%s\" [shape=note];\n", node))
		b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", node, r.Target))
	}

	b.WriteString("}\n")
	return b.String()
}

func buildScaffoldArgs(g graph) []string {
	var args []string
	for _, c := range g.Concepts {
		args = append(args, "--concept", c.Name)
		for _, f := range c.Fields {
			args = append(args, "--field", fmt.Sprintf("%s:%s:%s", c.Name, f.Name, f.Type))
		}
		for _, m := range c.Methods {
			args = append(args, "--method", fmt.Sprintf("%s:%s", c.Name, m))
		}
	}
	for _, o := range g.Orchestrators {
		args = append(args, "--orchestrator", o.Name)
		for _, p := range o.Params {
			args = append(args, "--param", fmt.Sprintf("%s:%s:%s", o.Name, p.Name, p.Type))
		}
		if o.Route != nil {
			args = append(args, "--route", fmt.Sprintf("%s:%s:%s", o.Route.Method, o.Route.Path, o.Route.Target))
		}
	}
	for _, p := range g.Projections {
		args = append(args, "--projection", p.Name)
		for _, q := range p.Query {
			args = append(args, "--query", fmt.Sprintf("%s:%s:%s", p.Name, q.Name, q.Type))
		}
		for _, r := range p.Result {
			args = append(args, "--result", fmt.Sprintf("%s:%s:%s", p.Name, r.Name, r.Type))
		}
		if p.Route != nil {
			args = append(args, "--route", fmt.Sprintf("%s:%s:%s", p.Route.Method, p.Route.Path, p.Route.Target))
		}
	}

	return args
}

func collectRoutes(orchestrators []orchestrator, projections []projection) []route {
	var routes []route
	for _, o := range orchestrators {
		if o.Route != nil {
			routes = append(routes, *o.Route)
		}
	}
	for _, p := range projections {
		if p.Route != nil {
			routes = append(routes, *p.Route)
		}
	}
	return routes
}

func runScaffold(root, module string, args []string) error {
	if root == "" {
		root = "."
	}
	cmdArgs := []string{"run", "./tools/scaffold", "init", "--root", root}
	if module != "" {
		cmdArgs = append(cmdArgs, "--module", module)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func symbolify(value string) string {
	parts := splitWords(value)
	for i, part := range parts {
		if len(part) > 0 {
			runes := []rune(strings.ToLower(part))
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, "")
}

func splitWords(value string) []string {
	var words []string
	var current []rune
	for _, r := range value {
		if r == ' ' || r == '-' || r == '_' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
			continue
		}
		current = append(current, r)
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}

func similarity(a, b string) float64 {
	if a == b {
		return 1
	}
	la := strings.ToLower(a)
	lb := strings.ToLower(b)
	dist := levenshtein(la, lb)
	maxLen := len(la)
	if len(lb) > maxLen {
		maxLen = len(lb)
	}
	if maxLen == 0 {
		return 1
	}
	return 1 - float64(dist)/float64(maxLen)
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min(del, ins, sub)
		}
		copy(prev, curr)
	}
	return prev[len(b)]
}

func conceptNames(concepts []concept) []string {
	var names []string
	for _, c := range concepts {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	return names
}

func orchestratorNames(orchestrators []orchestrator) []string {
	var names []string
	for _, o := range orchestrators {
		names = append(names, o.Name)
	}
	sort.Strings(names)
	return names
}

func projectionNames(projections []projection) []string {
	var names []string
	for _, p := range projections {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return names
}
