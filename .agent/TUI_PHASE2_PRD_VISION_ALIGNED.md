# Claudicus Phase 2 - Product Requirements Document

## Executive Summary

Phase 2 demonstrates Claudicus's core value proposition: **the unique combination of Uzi's operational excellence with Claude Squad's user experience excellence**. This phase doesn't aim to complete an MVP - it aims to prove that this fusion creates capabilities neither parent system can provide alone.

**Mission**: Show users a compelling glimpse of the Claudicus vision through one breakthrough capability that showcases operational speed + visual richness working together.

## Phase 2 Vision Alignment

### What We're Proving

**Core Hypothesis**: "Combining Uzi's operational backbone with Claude Squad's visual excellence creates unique value that neither system can achieve alone."

**Success Indicator**: Users say *"I can see why this approach is different - this gives me capabilities I couldn't get from either Uzi or Claude Squad individually."*

### What We're NOT Trying to Prove

- That Claudicus is faster than pure Uzi (Phase 2 won't be)
- That Claudicus is more polished than Claude Squad (Phase 2 won't be)  
- That Claudicus is a complete replacement for either system (it's not ready)

### What We ARE Trying to Prove

- That operational excellence + visual excellence = unique capabilities
- That the vision is compelling and worth pursuing
- That users can follow the roadmap to a superior future state

## Product Vision Reinforcement

**From PRODUCT_VISION.md**: "Claudicus combines Uzi's operational excellence with Claude Squad's user experience mastery to create the definitive platform for safe, multi-agent software development."

**Phase 2 Contribution**: Demonstrate the first concrete example of this fusion creating unique value.

---

## Core Feature: Visual Agent Coordination Hub

### The Breakthrough Capability

**Feature**: Real-time git diff preview with broadcast coordination

**Why This Showcases the Vision**:

- **Uzi DNA**: Leverages existing reliable git worktree system and broadcast commands
- **Claude Squad DNA**: Rich visual display that makes coordination intuitive and beautiful
- **Unique Value**: Capability impossible in either parent system alone

### User Story (Vision-Aligned)

**As a** developer managing multiple AI agents  
**I want to** see what each agent is working on AND coordinate them visually  
**So that** I can orchestrate my agents like a conductor with an orchestra, not like sending blind commands into the void

**Acceptance Criteria (Arrange-Act-Assert Format):**

```gherkin
Scenario: Visual coordination that neither parent system can provide
  Given I have 3 active agents working on different parts of my codebase
  When I open Claudicus TUI
  Then I can see all agents in the main list (Phase 1 capability)
  When I press Tab to enter diff preview mode
  Then I see a split view showing selected agent's git changes in real-time
  When I navigate between agents (up/down arrows)
  Then the diff preview updates instantly to show each agent's current changes
  When I press 'b' for broadcast
  Then I see a simple prompt "Message: " at the bottom
  When I type "Please add error handling" and press Enter
  Then I see "Broadcasting..." status message
  And I can watch each agent's status update as they receive and process the message
  And the diff preview continues updating to show their responses
```

```gherkin
Scenario: Operational reliability meets visual richness
  Given I'm using the diff preview functionality
  When the git command fails or returns invalid data
  Then I see a clear error message in the preview pane
  And the TUI remains stable and responsive
  And I can continue using other functionality normally
  When git operations succeed
  Then diff updates are smooth and non-disruptive
  And the visual display enhances rather than slows my workflow
```

### Technical Implementation

#### Core Components

**1. Git Diff Preview Pane**

- **Tab key**: Toggle between list view and split view (list + diff preview)
- **Arrow navigation**: Updates diff preview in real-time as you select different agents
- **Reliable backend**: Uses Uzi's proven git worktree system
- **Rich display**: Syntax-highlighted, properly formatted git diff output

**2. Simple Broadcast Integration**

- **'b' key**: Show simple prompt at bottom of screen
- **Command execution**: Leverages existing `uzi broadcast` command
- **Visual feedback**: Show status updates as agents receive and process message
- **Real-time coordination**: Watch agent status change in main list view

**3. Enhanced Status Display**

- **Better visual indicators**: Show when agents are processing broadcasts
- **Activity feedback**: Visual cues when agents are actively working
- **Status clarity**: Clear differentiation between idle, working, and processing states

#### Implementation Philosophy (Vision-Aligned)

**Operational Excellence (Uzi DNA)**:

- Use existing `uzi broadcast` command - don't reimplement
- Leverage proven git worktree system for diff generation
- Simple, reliable command execution
- Trust existing error handling patterns

**User Experience Excellence (Claude Squad DNA)**:

- Smooth, real-time visual updates
- Beautiful syntax highlighting and diff formatting
- Intuitive keyboard navigation
- Professional, polished visual design

**Unique Fusion Value**:

- Real-time coordination visibility impossible with CLI
- Visual context that enhances operational speed
- Orchestration capabilities neither parent provides

### Scope Boundaries

#### What's In Scope

- Git diff preview with Tab navigation
- Simple broadcast with visual status feedback
- Enhanced agent status indicators
- Reliable error handling and display

#### What's Out of Scope (Future Phases)

- Complex overlay input systems
- Multiple hotkey combinations (n, k, c, r)
- Advanced error recovery mechanisms
- Checkpoint management
- Help systems and advanced UI features

#### Technical Requirements

**Performance Standards**:

- Diff preview updates: <200ms (visual richness with acceptable speed)
- Broadcast execution: Leverage existing Uzi speed
- Status updates: Real-time without blocking navigation
- Memory usage: Stable during extended use

**Reliability Standards**:

- Failed git operations don't crash the TUI
- Broadcast failures show clear error messages
- Navigation remains responsive during all operations
- Terminal state properly restored on exit

**User Experience Standards**:

- Diff preview is syntax-highlighted and properly formatted
- Status updates are visually clear and non-disruptive
- Keyboard navigation feels smooth and predictable
- Visual design maintains Claude Squad quality standards

### Testing Requirements

#### Critical Path Testing (100% Coverage)

- Git diff generation and display logic
- Broadcast command execution and status tracking
- Navigation state management (Tab toggle, arrow selection)
- Error handling for git and broadcast operations

#### Integration Testing (All Scenarios)

- Complete user workflow from the acceptance criteria
- Error scenarios with graceful degradation
- Multiple agent coordination scenarios
- Extended use testing (memory and performance)

#### Vision Validation Testing

- User feedback sessions focused on "unique value" question
- Comparison testing against parent systems for vision clarity
- Workflow testing to validate operational + visual benefits

## Success Metrics

### Vision Alignment Metrics

**Primary Success**: Users understand and are excited about the Claudicus vision

- "I can see why this combination is powerful"
- "This gives me capabilities I couldn't get from Uzi or Claude Squad alone"
- "I want to see where this roadmap leads"

**Secondary Success**: Technical proof of concept validates the approach

- Git diff preview provides genuine operational value
- Broadcast coordination demonstrates visual orchestration benefits
- Performance and reliability meet acceptable standards

### User Feedback Targets

**Vision Clarity**: Users can articulate why Claudicus is different
**Value Recognition**: Users see concrete benefits from the fusion approach
**Future Interest**: Users want to continue using and following development

### Technical Success Criteria

**Operational Excellence**: Git and broadcast operations work reliably
**Visual Excellence**: Diff preview and status displays are clear and helpful
**Fusion Value**: The combination provides capabilities neither parent offers

## Implementation Approach

### Development Philosophy

**"Nail It Before Scale It" Applied**:

- Focus on one breakthrough capability rather than multiple features
- Implement it reliably rather than comprehensively
- Prove the vision works rather than building complete functionality

**Vision-Driven Development**:

- Every implementation decision must reinforce the fusion value proposition
- Prioritize capabilities that showcase both operational + visual excellence
- Accept limitations that don't compromise the core vision demonstration

### Phase Sequencing

**Phase 2.1: Git Diff Preview Foundation**

- Implement Tab toggle between list and split view
- Basic git diff display with navigation
- Reliable error handling and display

**Phase 2.2: Broadcast Integration**

- Add simple broadcast prompt and execution
- Visual status feedback for message processing
- Real-time status updates in main list

**Phase 2.3: Polish and Vision Validation**

- Enhance visual design and status indicators
- Performance optimization and testing
- User feedback collection and vision validation

### Risk Mitigation

**Technical Risks**:

- Git diff parsing failures → Clear error display, TUI remains stable
- Broadcast command failures → Leverage existing Uzi error handling
- Performance degradation → Profile and optimize critical paths

**Vision Risks**:

- Users don't see unique value → Gather feedback and iterate on presentation
- Fusion feels like mashup → Focus on seamless integration and visual cohesion
- Expectations exceed Phase 2 scope → Clear communication about roadmap

## Definition of Done

### Feature Completeness

- [ ] Tab navigation between list and diff preview views works smoothly
- [ ] Git diff preview shows real-time, syntax-highlighted changes for selected agent
- [ ] Broadcast functionality executes reliably with visual status feedback
- [ ] Enhanced status indicators clearly show agent activity states
- [ ] Error scenarios handled gracefully without crashing TUI

### Vision Validation

- [ ] Users can articulate why Claudicus approach is different/valuable
- [ ] Demonstration shows capabilities neither parent system provides
- [ ] User feedback indicates excitement about the vision and roadmap
- [ ] Technical implementation proves fusion approach is viable

### Quality Standards

- [ ] All critical functionality has 100% test coverage
- [ ] Performance meets acceptable standards for Phase 2 scope
- [ ] No regressions in Phase 1 functionality
- [ ] Documentation clearly explains vision alignment and future roadmap

### Success Validation

- [ ] Users say "I can see the unique value this combination provides"
- [ ] Technical proof of concept validates the operational + visual excellence fusion
- [ ] Clear path to future phases is established and validated

## Future Phases Preview

**Phase 3**: Advanced Coordination Features (batch operations, multi-agent selection)
**Phase 4**: Enhanced Lifecycle Management (create/kill with visual feedback)
**Phase 5**: Orchestration Automation (smart agent routing, dependency management)

**Each phase builds on the proven foundation**: Operational excellence + Visual excellence = Unique capabilities

---

**Document Version: 1.0**  
**Created: 2025-06-26**  
**Vision Alignment**: Demonstrates Claudicus fusion value without requiring MVP completion**  
**Success Measure**: Users see and want the future we're building**
