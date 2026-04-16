package configproviders

import (
	"fmt"
	"net"

	"github.com/hirochachacha/go-smb2"
)

type SMBInput struct {
	Addr     string
	User     string
	Password string
	Domain   string
}

type SMBManager struct {
	Addr     string
	User     string
	Password string
	Domain   string

	session *smb2.Session
	share   *smb2.Share
}

// addr - ip:port
func NewSMBManager(input SMBInput) (*SMBManager, error) {
	if input.Addr == "" {
		return nil, fmt.Errorf("addr smb is empty")
	}

	return &SMBManager{
		Addr:     input.Addr,
		User:     input.User,
		Password: input.Password,
		Domain:   input.Domain,
	}, nil
}

func (m *SMBManager) ConnectToShare() error {
	s, e := m.getSession()
	if e != nil {
		return e
	}

	fs, err := s.Mount("share")
	if err != nil {
		return err
	}

	m.share = fs
	return nil
}

func (m *SMBManager) GetShare() *smb2.Share {
	return m.share
}

func (m *SMBManager) Disconnect() {
	m.session.Logoff()
	m.share.Umount()
}

func (m *SMBManager) getSession() (*smb2.Session, error) {
	conn, err := net.Dial("tcp", m.Addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// auth
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     m.User,
			Password: m.Password,
			Domain:   m.Domain, // или WORKGROUP / DOMAIN
		},
	}

	// session
	s, err := d.Dial(conn)
	if err != nil {
		return nil, err
	}

	return s, nil
}
