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
	SchemaVersion    string         `json:"schema_version"`
	Environment      string         `json:"environment"`
	Query            string         `json:"query"`
	ID               string         `json:"id"`
	Kind             string         `json:"kind"`
	Found            bool           `json:"found"`
	Node             Node           `json:"node"`
	OutboundRefs     []string       `json:"outbound_refs"`
	InboundRefs      []string       `json:"inbound_refs"`
	Chains           [][]string     `json:"chains"`
	RelatedIntent    []EvidenceLink `json:"related_intent,omitempty"`
	RelatedBehavior  []EvidenceLink `json:"related_behavior,omitempty"`
	RelatedDesign    []EvidenceLink `json:"related_design,omitempty"`
	RelatedAssurance []EvidenceLink `json:"related_assurance,omitempty"`
	RelatedSecurity  []EvidenceLink `json:"related_security,omitempty"`
	RelatedExecution []EvidenceLink `json:"related_execution,omitempty"`
	MissingLinks     []string       `json:"missing_links,omitempty"`
	Recommendation   string         `json:"recommendation,omitempty"`
	Warnings         []string       `json:"warnings"`
	BrokenRefs       []string       `json:"broken_refs"`
}

type EvidenceLink struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	ArtifactPath string `json:"artifact_path"`
	Title        string `json:"title,omitempty"`
	Status       string `json:"status,omitempty"`
	Source       string `json:"source,omitempty"`
	Relation     string `json:"relation,omitempty"`
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
	related := g.relatedEvidence(id)
	missingLinks := g.missingLinks(id, related, warnings, brokenRefs)
	return TraceResult{
		SchemaVersion:    SchemaVersion,
		Environment:      g.Environment,
		Query:            id,
		ID:               id,
		Kind:             node.Type,
		Found:            true,
		Node:             node,
		OutboundRefs:     sortedStrings(node.Refs),
		InboundRefs:      g.inboundRefs(id),
		Chains:           g.chains(id),
		RelatedIntent:    related[TypeIntent],
		RelatedBehavior:  related[TypeBehavior],
		RelatedDesign:    related[TypeDesign],
		RelatedAssurance: related[TypeAssurance],
		RelatedSecurity:  related[TypeSecurity],
		RelatedExecution: related[TypeExecution],
		MissingLinks:     missingLinks,
		Recommendation:   recommendationForTrace(node, missingLinks),
		Warnings:         warnings,
		BrokenRefs:       brokenRefs,
	}, nil
}

func RenderText(result TraceResult) string {
	var lines []string
	lines = append(lines,
		fmt.Sprintf("Trace: %s", result.Query),
		"",
		result.Node.Type+":",
		fmt.Sprintf("- Found in %s", result.Node.ArtifactPath),
	)
	if result.Node.Title != "" {
		lines = append(lines, "- "+result.Node.Title)
	}
	if result.Node.Status != "" {
		lines = append(lines, "- Status: "+result.Node.Status)
	}
	if result.Node.Critical {
		lines = append(lines, "- Critical: true")
	}

	lines = appendTraceSection(lines, "Related intent:", result.RelatedIntent, emptyRelatedText(result, TypeIntent))
	lines = appendTraceSection(lines, "Related behavior:", result.RelatedBehavior, emptyRelatedText(result, TypeBehavior))
	lines = appendTraceSection(lines, "Design evidence:", result.RelatedDesign, "No design evidence references "+result.Query+".")
	lines = appendTraceSection(lines, "Assurance evidence:", result.RelatedAssurance, "No mapped test evidence references "+result.Query+".")
	lines = appendTraceSection(lines, "Security evidence:", result.RelatedSecurity, "No security evidence references "+result.Query+".")
	lines = appendTraceSection(lines, "Execution evidence:", result.RelatedExecution, "No telemetry or execution signal references "+result.Query+".")

	lines = append(lines, "", "Evidence Chain:")
	if len(result.Chains) == 0 {
		lines = append(lines, "- None.")
	} else {
		for _, chain := range result.Chains {
			lines = append(lines, strings.Join(chain, " -> "))
		}
	}
	lines = append(lines, "", "Missing links:")
	lines = appendList(lines, result.MissingLinks, "None.")
	if len(result.BrokenRefs) > 0 {
		lines = append(lines, "", "Broken references:")
		lines = appendList(lines, result.BrokenRefs, "None.")
	}
	lines = append(lines, "", "Recommendation:", result.Recommendation)

	return strings.Join(lines, "\n")
}

func RenderJSON(result TraceResult) (string, error) {
	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func appendTraceSection(lines []string, heading string, links []EvidenceLink, empty string) []string {
	lines = append(lines, "", heading)
	if len(links) == 0 {
		if strings.HasPrefix(empty, "No ") {
			return append(lines, "- Missing: "+empty)
		}
		return append(lines, "- "+empty)
	}
	for _, link := range links {
		text := fmt.Sprintf("- %s found in %s", link.ID, link.ArtifactPath)
		if link.Title != "" {
			text += " (" + link.Title + ")"
		}
		if link.Status != "" {
			text += " status=" + link.Status
		}
		if link.Relation != "" {
			text += " [" + link.Relation + "]"
		}
		lines = append(lines, text)
	}
	return lines
}

func emptyRelatedText(result TraceResult, sectionType string) string {
	if result.Node.Type == sectionType {
		return "None."
	}
	return "No " + strings.ToLower(sectionType) + " evidence references " + result.Query + "."
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
			if !graph.behaviorLinkedToAssurance(id) {
				if node.Critical {
					findings = append(findings, escalatedFinding(node, productionOrStrict, fmt.Sprintf("%s %s has no mapped test evidence", node.ArtifactPath, id)))
				} else {
					findings = append(findings, warningFinding(node, fmt.Sprintf("Traceability Gap: %s has no mapped test evidence", id)))
				}
			} else if !graph.behaviorHasAssuranceEvidence(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("Traceability Gap: %s has no mapped test evidence", id)))
			}
		case TypeAssurance:
			if !graph.assuranceLinkedToBehavior(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("%s %s is not linked to any behavior", node.ArtifactPath, id)))
			}
		case TypeSecurity:
			if !graph.securityLinkedToReleaseEvidence(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("%s %s is not mapped to behavior or intent evidence", node.ArtifactPath, id)))
			}
		case TypeExecution:
			if !graph.executionLinkedToBehaviorOrAssurance(id) {
				findings = append(findings, warningFinding(node, fmt.Sprintf("%s %s is not tied to behavior or assurance release readiness evidence", node.ArtifactPath, id)))
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

func (g Graph) InboundRefs(id string) []string {
	return g.inboundRefs(id)
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

func (g Graph) relatedEvidence(id string) map[string][]EvidenceLink {
	related := map[string][]EvidenceLink{
		TypeIntent:    {},
		TypeBehavior:  {},
		TypeDesign:    {},
		TypeAssurance: {},
		TypeSecurity:  {},
		TypeExecution: {},
	}
	query := g.Nodes[id]
	anchors := map[string]struct{}{id: {}}
	if query.Type == TypeIntent {
		for _, behaviorID := range g.relatedBehaviorIDs(id) {
			anchors[behaviorID] = struct{}{}
		}
	}
	if query.Type != TypeBehavior {
		for _, ref := range query.Refs {
			if typeForID(ref) == TypeBehavior {
				anchors[ref] = struct{}{}
			}
		}
		for _, inbound := range g.inboundRefs(id) {
			if typeForID(inbound) == TypeBehavior {
				anchors[inbound] = struct{}{}
			}
		}
	}

	for _, candidateID := range g.OrderedIDs {
		if candidateID == id {
			continue
		}
		candidate := g.Nodes[candidateID]
		relation := relationToAnchors(candidate, query, anchors)
		if relation == "" {
			continue
		}
		related[candidate.Type] = append(related[candidate.Type], evidenceLink(candidate, relation))
	}

	for key, links := range related {
		related[key] = uniqueLinks(links)
	}
	return related
}

func (g Graph) relatedBehaviorIDs(id string) []string {
	var ids []string
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeBehavior {
			ids = append(ids, ref)
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeBehavior {
			ids = append(ids, inbound)
		}
	}
	return sortedStrings(ids)
}

func relationToAnchors(candidate Node, query Node, anchors map[string]struct{}) string {
	if _, ok := anchors[candidate.ID]; ok {
		return "related behavior"
	}
	for _, ref := range candidate.Refs {
		if _, ok := anchors[ref]; ok {
			return "references " + ref
		}
		if ref == query.ID {
			return "references " + query.ID
		}
	}
	for anchor := range anchors {
		if containsString(query.Refs, candidate.ID) && anchor == query.ID {
			return "referenced by " + query.ID
		}
	}
	return ""
}

func evidenceLink(node Node, relation string) EvidenceLink {
	return EvidenceLink{
		ID:           node.ID,
		Kind:         node.Type,
		ArtifactPath: node.ArtifactPath,
		Title:        node.Title,
		Status:       node.Status,
		Source:       node.Source,
		Relation:     relation,
	}
}

func uniqueLinks(links []EvidenceLink) []EvidenceLink {
	seen := map[string]struct{}{}
	unique := make([]EvidenceLink, 0, len(links))
	for _, link := range links {
		if _, ok := seen[link.ID]; ok {
			continue
		}
		seen[link.ID] = struct{}{}
		unique = append(unique, link)
	}
	sort.Slice(unique, func(i, j int) bool {
		return nodeSortKey(unique[i].ID) < nodeSortKey(unique[j].ID)
	})
	return unique
}

func (g Graph) missingLinks(id string, related map[string][]EvidenceLink, warnings []string, broken []string) []string {
	node := g.Nodes[id]
	var missing []string
	switch node.Type {
	case TypeIntent:
		if len(related[TypeBehavior]) == 0 {
			missing = append(missing, id+" has no related behavior evidence")
		}
		for _, behavior := range related[TypeBehavior] {
			if !g.behaviorHasDesignEvidence(behavior.ID) {
				missing = append(missing, behavior.ID+" has no design reference")
			}
			if !g.behaviorHasAssuranceEvidence(behavior.ID) {
				missing = append(missing, behavior.ID+" has no mapped test evidence")
			}
		}
		if len(related[TypeExecution]) == 0 {
			missing = append(missing, "no execution signal references "+id)
		}
	case TypeBehavior:
		if len(related[TypeIntent]) == 0 {
			missing = append(missing, id+" has no related intent evidence")
		}
		if len(related[TypeDesign]) == 0 {
			missing = append(missing, id+" has no design reference")
		}
		if len(related[TypeAssurance]) == 0 {
			missing = append(missing, id+" has no mapped test evidence")
		}
		if len(related[TypeSecurity]) == 0 {
			missing = append(missing, id+" has no security evidence")
		}
		if len(related[TypeExecution]) == 0 {
			missing = append(missing, id+" has no execution signal")
		}
	}
	missing = append(missing, warnings...)
	missing = append(missing, broken...)
	return uniqueStrings(missing)
}

func recommendationForTrace(node Node, missing []string) string {
	if len(missing) == 0 {
		return "Keep related evidence current as this ID changes."
	}
	switch node.Type {
	case TypeIntent:
		return "Add or repair behavior and downstream evidence for " + node.ID + " so the intent can be audited end to end."
	case TypeBehavior:
		if missingAssuranceEvidence(missing) {
			if isPaymentRetryBehavior(node) {
				return "Add assurance evidence for payment retry behavior."
			}
			if onlyMissingAssuranceEvidence(missing) {
				return "Add mapped assurance evidence that references " + node.ID + "."
			}
		}
		return "Add design, assurance, security, and execution evidence that references " + node.ID + "."
	case TypeAssurance:
		return "Add refs from " + node.ID + " to the behavior it validates."
	case TypeSecurity:
		return "Add refs from " + node.ID + " to the intent or behavior affected by the security evidence."
	case TypeExecution:
		return "Add refs from " + node.ID + " to the behavior or assurance evidence it measures."
	default:
		return "Repair missing links for " + node.ID + "."
	}
}

func missingAssuranceEvidence(missing []string) bool {
	for _, item := range missing {
		lower := strings.ToLower(item)
		if strings.Contains(lower, "no mapped test evidence") ||
			strings.Contains(lower, "no assurance result") ||
			strings.Contains(lower, "not linked to assurance evidence") {
			return true
		}
	}
	return false
}

func isPaymentRetryBehavior(node Node) bool {
	lower := strings.ToLower(node.ID + " " + node.Title)
	return strings.Contains(lower, "behavior-003") ||
		(strings.Contains(lower, "retry") && strings.Contains(lower, "duplicate"))
}

func onlyMissingAssuranceEvidence(missing []string) bool {
	hasAssurance := false
	for _, item := range missing {
		lower := strings.ToLower(item)
		if strings.Contains(lower, "no mapped test evidence") ||
			strings.Contains(lower, "no assurance result") ||
			strings.Contains(lower, "not linked to assurance evidence") {
			hasAssurance = true
			continue
		}
		if strings.Contains(lower, "no design reference") ||
			strings.Contains(lower, "no security evidence") ||
			strings.Contains(lower, "no execution signal") ||
			strings.Contains(lower, "no related intent evidence") ||
			strings.Contains(lower, "not linked to intent evidence") {
			return false
		}
	}
	return hasAssurance
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

func (g Graph) behaviorHasAssuranceEvidence(id string) bool {
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeAssurance {
			return true
		}
	}
	return false
}

func (g Graph) behaviorHasDesignEvidence(id string) bool {
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeDesign {
			return true
		}
	}
	node := g.Nodes[id]
	for _, ref := range node.Refs {
		if typeForID(ref) == TypeDesign {
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
		if typeForID(ref) == TypeBehavior || typeForID(ref) == TypeIntent {
			return true
		}
	}
	for _, inbound := range g.inboundRefs(id) {
		if typeForID(inbound) == TypeBehavior || typeForID(inbound) == TypeIntent {
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
