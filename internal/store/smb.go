package store

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/hirochachacha/go-smb2"
)

// smbConnection holds SMB connection state
type smbConnection struct {
	conn    net.Conn
	session *smb2.Session
	share   *smb2.Share
}

// connectSMB establishes a connection to the SMB share using config
func connectSMB() (*smbConnection, error) {
	cfg := SMBConfig
	if !cfg.Enabled {
		LogDebug("SMB is disabled in config")
		return nil, nil
	}

	addr := cfg.Host + ":" + cfg.Port
	LogInfo("Connecting to SMB server: %s", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		LogError("Failed to connect to SMB server %s: %v", addr, err)
		return nil, fmt.Errorf("failed to connect to SMB server: %w", err)
	}
	LogDebug("TCP connection established")

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     cfg.User,
			Password: cfg.Password,
		},
	}

	session, err := d.Dial(conn)
	if err != nil {
		conn.Close()
		LogError("SMB authentication failed for user %s: %v", cfg.User, err)
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	LogDebug("SMB session established")

	share, err := session.Mount(cfg.Share)
	if err != nil {
		session.Logoff()
		conn.Close()
		LogError("Failed to mount SMB share %s: %v", cfg.Share, err)
		return nil, fmt.Errorf("failed to mount share: %w", err)
	}
	LogInfo("SMB share mounted: %s", cfg.Share)

	return &smbConnection{
		conn:    conn,
		session: session,
		share:   share,
	}, nil
}

// close closes the SMB connection
func (s *smbConnection) close() {
	if s == nil {
		return
	}
	LogDebug("Closing SMB connection")
	if s.share != nil {
		s.share.Umount()
	}
	if s.session != nil {
		s.session.Logoff()
	}
	if s.conn != nil {
		s.conn.Close()
	}
	LogInfo("SMB connection closed")
}

// syncToSMB copies the local database to the SMB share
func (s *smbConnection) syncToSMB() error {
	if s == nil || s.share == nil {
		LogDebug("syncToSMB: no SMB connection, skipping")
		return nil
	}

	localPath := GetDBPath()
	LogInfo("Syncing database to SMB: %s -> %s", localPath, DBFileName)

	localFile, err := os.Open(localPath)
	if err != nil {
		LogError("Failed to open local db for sync: %v", err)
		return fmt.Errorf("failed to open local db: %w", err)
	}
	defer localFile.Close()

	remoteFile, err := s.share.Create(DBFileName)
	if err != nil {
		LogError("Failed to create remote file: %v", err)
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	bytesWritten, err := io.Copy(remoteFile, localFile)
	if err != nil {
		LogError("Failed to copy to remote: %v", err)
		return fmt.Errorf("failed to copy to remote: %w", err)
	}

	LogInfo("SMB sync complete: %d bytes written", bytesWritten)
	return nil
}

// initSMB initializes SMB connection for the vault
func (v *FileVault) initSMB() {
	LogDebug("initSMB called, SMB enabled: %v", IsSMBEnabled())
	if !IsSMBEnabled() {
		LogInfo("SMB sync is disabled")
		return
	}

	smb, err := connectSMB()
	if err != nil {
		LogError("SMB connection failed: %v", err)
		return
	}
	if smb != nil {
		v.smb = smb
		LogInfo("SMB connection initialized successfully")
	}
}

// closeSMB closes the SMB connection
func (v *FileVault) closeSMB() {
	if v.smb != nil {
		v.smb.close()
		v.smb = nil
	}
}

// Sync copies the local database to the SMB share
func (v *FileVault) Sync() error {
	if v.smb == nil {
		LogDebug("Sync called but no SMB connection")
		return nil
	}
	return v.smb.syncToSMB()
}
