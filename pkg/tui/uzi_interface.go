// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"github.com/devflowinc/uzi/pkg/state"
)

// UziInterface defines the interface for interacting with Uzi core functionality
type UziInterface interface {
	// GetActiveSessions returns a list of active agent sessions
	GetActiveSessions() ([]string, error)
	
	// GetSessionState returns the state for a specific session
	GetSessionState(sessionName string) (*state.AgentState, error)
	
	// GetSessionStatus returns the current status of a session
	GetSessionStatus(sessionName string) (string, error)
	
	// AttachToSession attaches to an existing session
	AttachToSession(sessionName string) error
	
	// KillSession terminates a session
	KillSession(sessionName string) error
	
	// RefreshSessions refreshes the session list
	RefreshSessions() error
}

// UziClient implements the UziInterface for the TUI
type UziClient struct {
	stateManager *state.StateManager
}

// NewUziClient creates a new Uzi client for TUI operations
func NewUziClient() *UziClient {
	return &UziClient{
		stateManager: state.NewStateManager(),
	}
}

// GetActiveSessions implements UziInterface
func (c *UziClient) GetActiveSessions() ([]string, error) {
	// TODO: Implement session retrieval
	if c.stateManager == nil {
		return nil, nil
	}
	return c.stateManager.GetActiveSessionsForRepo()
}

// GetSessionState implements UziInterface
func (c *UziClient) GetSessionState(sessionName string) (*state.AgentState, error) {
	// TODO: Implement session state retrieval
	_ = sessionName
	return nil, nil
}

// GetSessionStatus implements UziInterface
func (c *UziClient) GetSessionStatus(sessionName string) (string, error) {
	// TODO: Implement session status retrieval
	_ = sessionName
	return "unknown", nil
}

// AttachToSession implements UziInterface
func (c *UziClient) AttachToSession(sessionName string) error {
	// TODO: Implement session attachment
	_ = sessionName
	return nil
}

// KillSession implements UziInterface
func (c *UziClient) KillSession(sessionName string) error {
	// TODO: Implement session termination
	_ = sessionName
	return nil
}

// RefreshSessions implements UziInterface
func (c *UziClient) RefreshSessions() error {
	// TODO: Implement session refresh
	return nil
}
