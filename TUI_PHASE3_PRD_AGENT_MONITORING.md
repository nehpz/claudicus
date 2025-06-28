# Claudicus Phase 3 - Product Requirements Document

## Executive Summary

Phase 3 transforms Claudicus from a passive session viewer into an **agent monitoring platform** for short-lived, parallel agent development. This phase focuses on solving the core problem users actually face: rapidly identifying which agents are producing valuable work versus which ones are stuck and should be replaced.

**Mission**: Enable users to efficiently monitor and manage parallel agent agents through real-time activity intelligence and rapid iteration workflows.

## Phase 3 Vision Alignment

### What We're Proving

**Core Hypothesis**: "Combining Uzi's operational reliability with Claude Squad's visual excellence creates unique agent monitoring capabilities that enable rapid parallel development iteration."

**Success Indicator**: Users say _"I can quickly see which of my 7 agents are actually working vs. stuck, preview their output without switching contexts, and rapidly kill/replace the duds."_

### What We're NOT Trying to Prove

- That Claudicus prevents file conflicts (conflicts are expected in parallel agents)
- That Claudicus coordinates agent work (agents should explore independently)
- That Claudicus suggests optimal agent counts (users know their problems better)

### What We ARE Trying to Prove

- That visual agent monitoring dramatically reduces time to identify productive vs. stuck agents
- That rapid kill/spawn workflows enable efficient agent iteration
- That activity intelligence helps users focus attention on promising approaches

## Product Vision Reinforcement

**From PRODUCT_VISION.md**: "Claudicus combines Uzi's operational excellence with Claude Squad's user experience mastery to create the definitive platform for safe, multi-agent software development."

**Phase 3 Contribution**: Demonstrate how this fusion enables **parallel agent workflows** that neither parent system can support.

---

## Core Feature: Agent Activity Intelligence

### The Breakthrough Capability

**Feature**: Real-time agent activity monitoring with rapid agent iteration

**Why This Showcases the Vision**:

- **Uzi DNA**: Leverages proven session management and git operations for reliable activity tracking
- **Claude Squad DNA**: Rich visual display that makes agent status immediately clear
- **Unique Value**: Parallel agent monitoring impossible with either CLI or single-agent TUI

### Key Workflow Patterns

**Context Reset vs. Prompt Refinement**:

- **Stuck strategy** (agent taking wrong approach) → Kill and respawn with **same prompt** for fresh context
- **Bad prompt** (unclear requirements) → Kill and respawn with **refined prompt** for better guidance
- **Working agents** → Continue running regardless of other agent status

**Worktree and Branch Management**:

- Kill operation **destroys tmux session + worktree + git branch** (including all commits)
- Replacement agent gets **fresh worktree + new branch** from main
- Failed exploration artifacts are **intentionally discarded** to prevent strategy contamination
- Successful agents preserve work for later cherry-picking

**Accidental Action Protection**:

- **'k' key requires confirmation** before destructive kill operation
- **Brief delay** before actual cleanup to allow cancellation
- **Clear visual feedback** distinguishing destructive vs. safe actions
- **Fat-finger protection** for valuable multi-hour work preservation

### User Story (Vision-Aligned)

**As a** developer running parallel agent agents
**I want to** monitor agent activity and quickly identify productive vs. stuck agents
**So that** I can efficiently iterate on failed agents and focus on promising approaches

**Acceptance Criteria (Arrange-Act-Assert Format):**

```gherkin
Scenario: Rapid agent assessment during 20-minute sprint
  Given I have 7 agents running parallel agents for 18 minutes
  When I view the agent monitoring display
  Then I see each agent's commit count, lines changed, and activity status
  And I can immediately identify that 2 agents are stuck with 0 commits
  And I can see that 2 agents are highly productive with 4+ commits
  When I select a stuck agent and press 'k'
  Then I see confirmation to kill and replace with new agent
  When I confirm the replacement
  Then the stuck agent is terminated and a new agent is spawned with a different approach
  And I can continue monitoring the updated agent set
```

```gherkin
Scenario: Quick agent preview without context switching
  Given I have multiple active agents with different approaches
  When I press Tab to cycle through agent previews
  Then I see each agent's recent commit message and file changes
  And I can assess their approach without opening tmux sessions
  And I can identify which agents are pursuing promising directions
  When I find an agent with substantial progress
  Then I can mark it for detailed review after the agent completes
```

```gherkin
Scenario: agent completion and result evaluation
  Given my 20-minute agent sprint is complete
  When I review the final agent statuses
  Then I can see which 2-3 agents produced substantial work
  And I can identify the 4-5 throwaway agents with minimal progress
  And I can quickly access the promising results for cherry-picking
  When I terminate the agent set
  Then all agents are cleanly shut down but worktrees remain for result extraction
```

### Technical Implementation

#### Core Components

**1. agent Activity Monitor**

- **Real-time commit tracking**: Monitor git log for each agent's worktree
- **Activity classification**: Categorize agents as "working", "idle", or "stuck"
- **Progress metrics**: Track commits, lines changed, files modified
- **Time tracking**: Monitor elapsed time and last activity timestamps

**2. Visual agent Dashboard**

- **Status overview**: Show all agents with activity indicators and progress bars
- **Quick metrics**: Display commit count, lines changed, last activity time
- **Visual alerts**: Highlight stuck agents that should be considered for replacement
- **Progress visualization**: Real-time updates without screen clearing

**3. Rapid Iteration Workflow**

- **'k' key**: Kill selected agent with replacement prompt
- **Tab navigation**: Cycle through agent previews without context switching
- **Status filtering**: Focus on productive vs. stuck agents
- **Batch operations**: Kill multiple stuck agents simultaneously

#### Implementation Philosophy (Vision-Aligned)

**Operational Excellence (Uzi DNA)**:

- Use existing git log and diff commands for activity tracking
- Leverage proven tmux session management for agent lifecycle
- Maintain Uzi's speed and reliability for kill/spawn operations
- Trust existing error handling and cleanup processes

**User Experience Excellence (Claude Squad DNA)**:

- Smooth, real-time visual updates showing agent progress
- Beautiful progress indicators and activity visualization
- Intuitive keyboard navigation for rapid agent management
- Professional, clear status displays that reduce cognitive load

**Unique Fusion Value**:

- Parallel agent monitoring impossible with pure CLI tools
- Visual activity intelligence that enhances operational speed
- Rapid iteration capabilities that neither parent system provides

### Scope Boundaries

#### What's In Scope

- Real-time activity monitoring for all active agents
- Visual agent status dashboard with progress indicators
- Kill and replace workflow for stuck agents
- Quick preview cycling through agent results
- Activity-based agent classification (working/idle/stuck)

#### What's Out of Scope (Future Phases)

- Automated agent quality scoring and review
- Automatic agent termination based on performance thresholds
- Complex agent orchestration or coordination
- Multi-project or team collaboration features
- Advanced code analysis or semantic understanding

#### Technical Requirements

**Performance Standards**:

- Activity updates: <500ms refresh cycle without blocking UI
- Agent status classification: Real-time without performance impact
- Kill/spawn operations: Maintain Uzi's operational speed
- Memory usage: Stable during extended agent sessions

**Reliability Standards**:

- Git operations work across different repository states
- Agent termination cleanly preserves worktree data
- Status monitoring continues during high agent activity
- Terminal state properly restored after agent completion

**User Experience Standards**:

- Activity status is immediately comprehensible
- Progress indicators provide genuine insight into agent productivity
- Navigation feels responsive during intense agent monitoring
- Visual design maintains Claude Squad quality standards

### Testing Requirements

#### Critical Path Testing (100% Coverage)

- Git activity monitoring and commit tracking logic
- Agent status classification algorithms
- Kill/spawn workflow execution and state management
- Activity display and visual update systems

#### Integration Testing (All Scenarios)

- Complete agent workflows from creation to completion
- Multiple agent coordination during intensive development periods
- Error scenarios with graceful degradation
- Extended agent sessions with resource management

#### Vision Validation Testing

- User feedback sessions focused on agent monitoring efficiency
- Time-to-identification testing for stuck vs. productive agents
- Workflow efficiency comparison against manual tmux monitoring

## Success Metrics

### Vision Alignment Metrics

**Primary Success**: Users can efficiently manage parallel agents

- "I can identify stuck agents in seconds instead of minutes"
- "I don't need to manually check tmux sessions to understand agent progress"
- "I can rapidly iterate on failed agents during development sprints"

**Secondary Success**: Technical foundation supports the agent workflow

- Activity monitoring provides accurate, actionable insights
- Kill/spawn operations maintain operational reliability
- Visual dashboard enhances rather than slows agent management

### User Feedback Targets

**Efficiency Gain**: Time to identify and replace stuck agents reduces from 5-10 minutes to <30 seconds
**Workflow Improvement**: Users prefer Claudicus agent monitoring over CLI-only approaches
**Iteration Speed**: Faster agent cycles enable more parallel approaches per development session

### Technical Success Criteria

**Operational Excellence**: Git monitoring and agent lifecycle operations work reliably
**Visual Excellence**: Activity displays provide clear, actionable agent insights
**Fusion Value**: The combination enables agent workflows impossible with either parent system

## Implementation Approach

### Development Philosophy

**"Nail It Before Scale It" Applied**:

- Focus on core agent monitoring rather than advanced automation
- Implement reliable activity tracking before sophisticated analysis
- Prove visual agent management works before building complex features

**Vision-Driven Development**:

- Every implementation decision must enable parallel agent workflows
- Prioritize capabilities that showcase operational + visual excellence fusion
- Accept limitations that don't compromise core agent monitoring value

### Phase Sequencing

**Phase 3.1: Activity Monitoring Foundation**

- Implement real-time git activity tracking
- Basic agent status classification (working/idle/stuck)
- Visual agent dashboard with progress indicators

**Phase 3.2: Rapid Iteration Workflow**

- Kill and replace functionality for stuck agents
- Quick preview cycling through agent results
- Enhanced status display with agent insights

**Phase 3.3: agent Management Polish**

- Performance optimization for high-activity agents
- Advanced visual indicators and progress tracking
- User experience refinement and workflow validation

### Risk Mitigation

**Technical Risks**:

- Git monitoring overhead → Efficient polling and caching strategies
- High agent activity impact → Asynchronous updates and resource management
- Kill/spawn reliability → Leverage existing Uzi operational patterns

**Vision Risks**:

- Users don't see agent monitoring value → Focus on clear productivity indicators
- Monitoring feels like overhead → Ensure visual feedback enhances rather than distracts
- agent workflow doesn't match user patterns → Gather feedback and iterate quickly

## Definition of Done

### Feature Completeness

- [ ] Real-time activity monitoring tracks all agent agents accurately
- [ ] Visual dashboard shows agent status, progress, and activity patterns
- [ ] Kill/replace workflow enables rapid agent iteration
- [ ] Quick preview cycling provides insight without context switching
- [ ] Agent classification accurately identifies productive vs. stuck agents

### Vision Validation

- [ ] Users demonstrate improved agent management efficiency
- [ ] Monitoring capabilities provide clear advantages over CLI-only workflows
- [ ] Technical implementation proves operational + visual excellence fusion value
- [ ] Clear foundation established for future agent automation features

### Quality Standards

- [ ] All critical agent workflows have 100% test coverage
- [ ] Performance meets standards during intensive parallel development
- [ ] No regressions in Phase 1-2 functionality
- [ ] Documentation clearly explains agent workflow benefits

### Success Validation

- [ ] Users say "Claudicus makes parallel agentation practical and efficient"
- [ ] Time to manage agents measurably improves over manual approaches
- [ ] Technical proof of concept validates agent monitoring as unique Claudicus value

## Future Phases Preview

**Phase 4**: Automated agent Quality Assessment (review agents, scoring thresholds)
**Phase 5**: Intelligent agent Orchestration (auto-spawn, adaptive prompts)
**Phase 6**: agent Result Integration (automated cherry-picking, conflict resolution)

**Each phase builds on proven agent monitoring**: Operational excellence + Visual excellence = Advanced automation capabilities

---

**Document Version: 1.0**
**Created: 2025-06-26**
**Vision Alignment**: Enables parallel agent workflows unique to Claudicus\*\*
**Success Measure**: Users efficiently manage short-sprint agent agentation\*\*
