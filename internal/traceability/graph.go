package traceability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	SchemaVersion = "trace.v1"

	SeverityFail    = "FAIL"
	SeverityWarning = "WARNING"

	TypeIntent    = "Intent"
	TypeBehavior  = "Behavior"
	TypeDesign    = "Design"
	TypeAssurance = "Assurance"
	TypeSecurity  = "Security"
	TypeExecution = "Execution"
)

var (
	validIDPattern         = regexp.MustCompile(`^(INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}$`)
	possibleIDPattern      = regexp.MustCompile(`^[A-Z]+-[A-Za-z0-9-]+$`)
	markdownHeadingPattern = regexp.MustCompile(`^\s{0,3}(#{1,6})\s+(.+?)\s*$`)
	markdownIDPattern      = regexp.MustCompile(`^([A-Z]+-[A-Za-z0-9-]+)(?::\s*(.*))?$`)
)

type Node struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Title        string   `json:"title,omitempty"`
	ArtifactPath string   `json:"artifact_path"`
	Critical     bool     `json:"critical,omitempty"`
	Refs         []string `json:"refs,omitempty"`
	Status       string   `json:"status,omitempty"`
	Source       string   `json:"source,omitempty"`
}

type Graph struct {
	Environment string
	Nodes       map[string]Node
	OrderedIDs  []string
	Findings    []Finding
}

type Finding struct {
	Severity     string
	ArtifactPath string
	SourceID     string
	ReferenceID  string
	Message      string
}

type Options struct {
	Environment string
	Strict      bool
}

type TraceResult struct {
	SchemaVersion string     `json:"schema_version"`
	Environment   string     `json:"environment"`
	Query         string     `json:"query"`
	Node          Node       `json:"node"`
	OutboundRefs  []string   `json:"outbound_refs"`
	InboundRefs   []string   `json:"inbound_refs"`
	Chains        [][]string `json:"chains"`
	Warnings      []string   `json:"warnings"`
	BrokenRefs    []string   `json:"broken_refs"`
}

type jsonEvidenceFile struct {
	Evidence []jsonEvidence `json:"evidence"`
}

type jsonEvidence struct {
	ID     string   `json:"id"`
	Refs   []string `json:"refs"`
	Source string   `json:"source"`
	Status string   `json:"status"`
	Title  string   `json:"title"`
}

type artifact struct {
	relativePath string
	nodeType     string
}

var markdownArtifacts = []artifact{
	{relativePath: "intent/intent.md", nodeType: TypeIntent},
	{relativePath: "behavior/behavior-spec.md", nodeType: TypeBehavior},
	{relativePath: "design/architecture.md", nodeType: TypeDesign},
}

var jsonArtifacts = []artifact{
	{relativePath: "assurance/results.json", nodeType: TypeAssurance},
	{relativePath: "security/guardrails.json", nodeType: TypeSecurity},
	{relativePath: "execution/telemetry.json", nodeType: TypeExecution},
}

func Build(rootPath string, options Options) (Graph, error) {
	graph := Graph{
		Environment: options.Environment,
		Nodes:       map[string]Node{},
	}

	for _, artifact := range markdownArtifacts {
		if err := parseMarkdownArtifact(rootPath, artifact, &graph); err != nil {
			return Graph{}, err
		}
	}
	for _, artifact := range jsonArtifacts {
		if err := parseJSONArtifact(rootPath, artifact, &graph); err != nil {
			return Graph{}, err
		}
	}

	graph.sortNodes()
	graph.Findings = append(graph.Findings, analyzeGraph(graph, options)...)
	return graph, nil
}

func (g Graph) ValidateFindings() []Finding {
	findings := append([]Finding{}, g.Findings...)
	sort.SliceStable(findings, func(i int, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity < findings[j].Severity
		}
		if findings[i].ArtifactPath != findings[j].ArtifactPath {
			return findings[i].ArtifactPath < findings[j].ArtifactPath
		}
		if findings[i].SourceID != findings[j].SourceID {
			return findings[i].SourceID < findings[j].SourceID
		}
		return findings[i].Message < findings[j].Message
	})
	return findings
}

func (g Graph) Trace(id string) (TraceResult, error) {
	node, ok := g.Nodes[id]
	if !ok {
		return TraceResult{}, fmt.Errorf("unknown trace id %q", id)
	}

	warnings, brokenRefs := g.traceMessages(id)
	return TraceResult{
		SchemaVersion: SchemaVersion,
		Environment:   g.Environment,
		Query:         id,
		Node:          node,
		OutboundRefs:  sortedStrings(node.Refs),
		InboundRefs:   g.inboundRefs(id),
		Chains:        g.chains(id),
		Warnings:      warnings,
		BrokenRefs:    brokenRefs,
	}, nil
}

func RenderText(result TraceResult) string {
	var lines []string
	lines = append(lines,
		fmt.Sprintf("Trace: %s", result.Query),
		fmt.Sprintf("Type: %s", result.Node.Type),
		fmt.Sprintf("Artifact: %s", result.Node.ArtifactPath),
	)
	if result.Node.Title != "" {
		lines = append(lines, fmt.Sprintf("Title: %s", result.Node.Title))
	}
	if result.Node.Status != "" {
		lines = append(lines, fmt.Sprintf("Status: %s", result.Node.Status))
	}
	if result.Node.Critical {
		lines = append(lines, "Critical: true")
	}

	lines = append(lines, "", "Outbound References:")
	lines = appendList(lines, result.OutboundRefs, "None.")
	lines = append(lines, "", "Inbound References:")
	lines = appendList(lines, result.InboundRefs, "None.")
	lines = append(lines, "", "Evidence Chain:")
	if len(result.Chains) == 0 {
		lines = append(lines, "- None.")
	} else {
		for _, chain := range result.Chains {
			lines = append(lines, strings.Join(chain, " -> "))
		}
	}
	lines = append(lines, "", "Warnings:")
	lines = appendList(lines, result.Warnings, "None.")
	lines = append(lines, "", "Broken References:")
	lines = appendList(lines, result.BrokenRefs, "None.")

	return strings.Join(lines, "\n")
}

func RenderJSON(result TraceResult) (string, error) {
	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func parseMarkdownArtifact(rootPath string, artifact artifact, graph *Graph) error {
	path := filepath.Join(rootPath, artifact.relativePath)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	for index := 0; index < len(lines); index++ {
		level, headingText, ok := parseMarkdownHeading(lines[index])
		if !ok {
			continue
		}

		id, title, found, valid := parseHeadingID(headingText)
		if !found {
			continue
		}
		if !valid {
			graph.addFinding(Finding{
				Severity:     SeverityFail,
				ArtifactPath: artifactPath(rootPath, artifact.relativePath),
				SourceID:     id,
				Message:      fmt.Sprintf("%s has invalid evidence ID %s", artifactPath(rootPath, artifact.relativePath), id),
			})
			continue
		}

		body := markdownBody(lines[index+1:], level)
		node := Node{
			ID:           id,
			Type:         typeForID(id),
			Title:        title,
			ArtifactPath: artifactPath(rootPath, artifact.relativePath),
			Critical:     parseCritical(body),
			Refs:         parseMarkdownRefs(body, artifactPath(rootPath, artifact.relativePath), id, graph),
		}
		graph.addNode(node)
	}

	return nil
}

func parseJSONArtifact(rootPath string, artifact artifact, graph *Graph) error {
	path := filepath.Join(rootPath, artifact.relativePath)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var data jsonEvidenceFile
	if err := json.Unmarshal(content, &data); err != nil {
		return nil
	}

	for _, evidence := range data.Evidence {
		if evidence.ID == "" {
			continue
		}
		if !validIDPattern.MatchString(evidence.ID) {
			graph.addFinding(Finding{
				Severity:     SeverityFail,
				ArtifactPath: artifactPath(rootPath, artifact.relativePath),
				SourceID:     evidence.ID,
				Message:      fmt.Sprintf("%s has invalid evidence ID %s", artifactPath(rootPath, artifact.relativePath), evidence.ID),
			})
			continue
		}

		refs := make([]string, 0, len(evidence.Refs))
		for _, ref := range evidence.Refs {
			if !validIDPattern.MatchString(ref) {
				graph.addFinding(Finding{
					Severity:     SeverityFail,
					ArtifactPath: artifactPath(rootPath, artifact.relativePath),
					SourceID:     evidence.ID,
					ReferenceID:  ref,
					Message:      fmt.Sprintf("%s %s has invalid reference %s", artifactPath(rootPath, artifact.relativePath), evidence.ID, ref),
				})
				continue
			}
			refs = append(refs, ref)
		}

		title := evidence.Title
		if title == "" {
			title = evidence.Source
		}
		node := Node{
			ID:           evidence.ID,
			Type:         typeForID(evidence.ID),
			Title:        title,
			ArtifactPath: artifactPath(rootPath, artifact.relativePath),
			Refs:         refs,
			Status:       evidence.Status,
			Source:       evidence.Source,
		}
		graph.addNode(node)
	}

	return nil
}

func analyzeGraph(graph Graph, options Options) []Finding {
	var findings []Finding

	for _, id := range graph.OrderedIDs {
		node := graph.Nodes[id]
		for _, ref := range node.Refs {
			if _, ok := graph.Nodes[ref]; !ok {
				findings = append(findings, Finding{
					Severity:     SeverityFail,
					ArtifactPath: node.ArtifactPath,
					SourceID:     node.ID,
					ReferenceID:  ref,
					Message:      fmt.Sprintf("%s %s references missing %s", node.ArtifactPath, node.ID, ref),
				})
				continue
			}
			if !referenceAllowed(node.ID, ref) {
				findings = append(findings, Finding{
					Severity:     SeverityFail,
					ArtifactPath: node.ArtifactPath,
					SourceID:     node.ID,
					ReferenceID:  ref,
					Message:      fmt.Sprintf("%s %s cannot reference %s", node.ArtifactPath, node.ID, ref),
				})
			}
		}
	}

	productionOrStrict := options.Strict || strings.EqualFold(options.Environment, "production")
	for _, id := range graph.OrderedIDs {
		node := graph.Nodes[id]
		switch node.Type {
		case TypeIntent:
			if !graph.intentLinkedToBehavior(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("%s %s is not linked to any behavior", node.ArtifactPath, id)))
			}
		case TypeBehavior:
			if !graph.behaviorLinkedToIntent(id) {
				findings = append(findings, escalatedFinding(node, productionOrStrict, fmt.Sprintf("%s %s is not linked to intent evidence", node.ArtifactPath, id)))
			}
			if node.Critical && !graph.behaviorLinkedToAssurance(id) {
				findings = append(findings, escalatedFinding(node, productionOrStrict, fmt.Sprintf("%s %s is critical but is not linked to assurance evidence", node.ArtifactPath, id)))
			}
		case TypeAssurance:
			if !graph.assuranceLinkedToBehavior(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("%s %s is not linked to any behavior", node.ArtifactPath, id)))
			}
		case TypeSecurity:
			if !graph.securityLinkedToReleaseEvidence(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("%s %s is not linked to behavior or assurance evidence", node.ArtifactPath, id)))
			}
		case TypeExecution:
			if !graph.executionLinkedToBehaviorOrAssurance(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("%s %s is not linked to behavior or assurance evidence", node.ArtifactPath, id)))
			}
		}
	}

	return findings
}

func (g *Graph) addNode(node Node) {
	if existing, ok := g.Nodes[node.ID]; ok {
		g.addFinding(Finding{
			Severity:     SeverityFail,
			ArtifactPath: node.ArtifactPath,
			SourceID:     node.ID,
			Message:      fmt.Sprintf("%s duplicates evidence ID %s first defined in %s", node.ArtifactPath, node.ID, existing.ArtifactPath),
		})
		return
	}

	node.Refs = sortedStrings(uniqueStrings(node.Refs))
	g.Nodes[node.ID] = node
	g.OrderedIDs = append(g.OrderedIDs, node.ID)
}

func (g *Graph) addFinding(finding Finding) {
	g.Findings = append(g.Findings, finding)
}

func (g *Graph) sortNodes() {
	sort.Slice(g.OrderedIDs, func(i int, j int) bool {
		return nodeSortKey(g.OrderedIDs[i]) < nodeSortKey(g.OrderedIDs[j])
	})
}

func (g Graph) inboundRefs(id string) []string {
	var inbound []string
	for _, candidateID := range g.OrderedIDs {
		if containsString(g.Nodes[candidateID].Refs, id) {
			inbound = append(inbound, candidateID)
		}
	}
	return inbound
}

func (g Graph) chains(id string) [][]string {
	visited := map[string]bool{}
	var component []string
	var walk func(string)
	walk = func(current string) {
		if visited[current] {
			return
		}
		visited[current] = true
		component = append(component, current)
		for _, ref := range g.Nodes[current].Refs {
			if _, ok := g.Nodes[ref]; ok {
				walk(ref)
			}
		}
		for _, inbound := range g.inboundRefs(current) {
			walk(inbound)
		}
	}
	walk(id)
	if g.Nodes[id].Type != TypeDesign {
		component = withoutType(component, TypeDesign)
	}
	sort.Slice(component, func(i int, j int) bool {
		return nodeSortKey(component[i]) < nodeSortKey(component[j])
	})
	if len(component) == 0 {
		return nil
	}
	return [][]string{component}
}

func withoutType(ids []string, nodeType string) []string {
	filtered := make([]string, 0, len(ids))
	for _, id := range ids {
		if typeForID(id) == nodeType {
			continue
		}
		filtered = append(filtered, id)
	}
	return filtered
}

func (g Graph) traceMessages(id string) ([]string, []string) {
	var warnings []string
	var broken []string
	for _, finding := range g.ValidateFindings() {
		if finding.SourceID != id && finding.ReferenceID != id {
			continue
		}
		switch finding.Severity {
		case SeverityFail:
			broken = append(broken, finding.Message)
		case SeverityWarning:
			warnings = append(warnings, finding.Message)
		}
	}

	node := g.Nodes[id]
	if node.Type == TypeBehavior && !g.behaviorLinkedToSecurity(id) {
		warnings = append(warnings, fmt.Sprintf("No security evidence linked to %s", id))
	}
	return uniqueStrings(warnings), uniqueStrings(broken)
}

func (g Graph) intentLinkedToBehavior(id string) bool {
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeBehavior {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeBehavior {
			return true
		}
	}
	return false
}

func (g Graph) behaviorLinkedToIntent(id string) bool {
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeIntent {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeIntent {
			return true
		}
	}
	return false
}

func (g Graph) behaviorLinkedToAssurance(id string) bool {
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeAssurance {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeAssurance {
			return true
		}
	}
	return false
}

func (g Graph) behaviorLinkedToSecurity(id string) bool {
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeSecurity {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeSecurity {
			return true
		}
	}
	return false
}

func (g Graph) assuranceLinkedToBehavior(id string) bool {
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeBehavior {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeBehavior {
			return true
		}
	}
	return false
}

func (g Graph) securityLinkedToReleaseEvidence(id string) bool {
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeBehavior || typeForID(ref) == TypeAssurance {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeBehavior || typeForID(inbound) == TypeAssurance {
			return true
		}
	}
	return false
}

func (g Graph) executionLinkedToBehaviorOrAssurance(id string) bool {
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeBehavior || typeForID(ref) == TypeAssurance {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeBehavior || typeForID(inbound) == TypeAssurance {
			return true
		}
	}
	return false
}

func referenceAllowed(sourceID string, targetID string) bool {
	sourceType := typeForID(sourceID)
	targetType := typeForID(targetID)

	allowed := map[string]map[string]bool{
		TypeIntent: {
			TypeBehavior: true,
		},
		TypeBehavior: {
			TypeIntent:    true,
			TypeAssurance: true,
		},
		TypeDesign: {
			TypeIntent:   true,
			TypeBehavior: true,
		},
		TypeAssurance: {
			TypeBehavior: true,
		},
		TypeSecurity: {
			TypeIntent:    true,
			TypeBehavior:  true,
			TypeAssurance: true,
			TypeExecution: true,
		},
		TypeExecution: {
			TypeBehavior:  true,
			TypeAssurance: true,
		},
	}

	return allowed[sourceType][targetType]
}

func parseMarkdownHeading(line string) (int, string, bool) {
	matches := markdownHeadingPattern.FindStringSubmatch(line)
	if matches == nil {
		return 0, "", false
	}
	level := len(matches[1])
	text := strings.TrimSpace(strings.TrimRight(matches[2], "#"))
	return level, text, true
}

func parseHeadingID(text string) (string, string, bool, bool) {
	matches := markdownIDPattern.FindStringSubmatch(text)
	if matches == nil {
		return "", "", false, false
	}
	id := matches[1]
	title := strings.TrimSpace(matches[2])
	if !possibleIDPattern.MatchString(id) {
		return "", "", false, false
	}
	return id, title, true, validIDPattern.MatchString(id)
}

func markdownBody(lines []string, currentLevel int) []string {
	var body []string
	for _, line := range lines {
		level, _, ok := parseMarkdownHeading(line)
		if ok && level <= currentLevel {
			break
		}
		body = append(body, line)
	}
	return body
}

func parseCritical(lines []string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, "Critical: true") {
			return true
		}
	}
	return false
}

func parseMarkdownRefs(lines []string, artifactPath string, sourceID string, graph *Graph) []string {
	var refs []string
	inRefs := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.EqualFold(trimmed, "Refs:") || strings.EqualFold(trimmed, "References:"):
			inRefs = true
			continue
		case inRefs && strings.HasPrefix(trimmed, "- "):
			fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			if len(fields) == 0 {
				continue
			}
			ref := fields[0]
			if !validIDPattern.MatchString(ref) {
				graph.addFinding(Finding{
					Severity:     SeverityFail,
					ArtifactPath: artifactPath,
					SourceID:     sourceID,
					ReferenceID:  ref,
					Message:      fmt.Sprintf("%s %s has invalid reference %s", artifactPath, sourceID, ref),
				})
				continue
			}
			refs = append(refs, ref)
		case inRefs && trimmed == "":
			continue
		case inRefs:
			inRefs = false
		}
	}
	return refs
}

func typeForID(id string) string {
	prefix := strings.SplitN(id, "-", 2)[0]
	switch prefix {
	case "INTENT":
		return TypeIntent
	case "BEHAVIOR":
		return TypeBehavior
	case "DESIGN":
		return TypeDesign
	case "ASSURANCE":
		return TypeAssurance
	case "SECURITY":
		return TypeSecurity
	case "EXECUTION":
		return TypeExecution
	default:
		return ""
	}
}

func artifactPath(rootPath string, relativePath string) string {
	return filepath.ToSlash(filepath.Join(filepath.Base(rootPath), relativePath))
}

func warningFinding(node Node, message string) Finding {
	return Finding{
		Severity:     SeverityWarning,
		ArtifactPath: node.ArtifactPath,
		SourceID:     node.ID,
		Message:      message,
	}
}

func escalatedFinding(node Node, fail bool, message string) Finding {
	severity := SeverityWarning
	if fail {
		severity = SeverityFail
	}
	return Finding{
		Severity:     severity,
		ArtifactPath: node.ArtifactPath,
		SourceID:     node.ID,
		Message:      message,
	}
}

func nodeSortKey(id string) string {
	return fmt.Sprintf("%02d:%s", typeOrder(typeForID(id)), id)
}

func typeOrder(nodeType string) int {
	switch nodeType {
	case TypeIntent:
		return 1
	case TypeBehavior:
		return 2
	case TypeDesign:
		return 3
	case TypeAssurance:
		return 4
	case TypeSecurity:
		return 5
	case TypeExecution:
		return 6
	default:
		return 99
	}
}

func existingRefs(refs []string, nodes map[string]Node) []string {
	var existing []string
	for _, ref := range refs {
		if _, ok := nodes[ref]; ok {
			existing = append(existing, ref)
		}
	}
	return existing
}

func appendList(lines []string, items []string, emptyText string) []string {
	if len(items) == 0 {
		return append(lines, "- "+emptyText)
	}
	for _, item := range items {
		lines = append(lines, "- "+item)
	}
	return lines
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func sortedStrings(values []string) []string {
	values = uniqueStrings(values)
	sort.Slice(values, func(i int, j int) bool {
		return nodeSortKey(values[i]) < nodeSortKey(values[j])
	})
	return values
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
