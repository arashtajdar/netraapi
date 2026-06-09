package services

import (
	"context"
	"testing"

	"sheedbox-api/models"
)

type mockLiveTVRepo struct {
	channels []models.LiveTVChannel
	epgMap   map[int][]models.EPG // channelID -> EPG entries
}

func (m *mockLiveTVRepo) List(ctx context.Context) ([]models.LiveTVChannel, error) {
	return m.channels, nil
}

func (m *mockLiveTVRepo) GetByID(ctx context.Context, id int) (*models.LiveTVChannel, error) {
	for _, c := range m.channels {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

func (m *mockLiveTVRepo) GetEPGForChannel(ctx context.Context, channelID int) ([]models.EPG, error) {
	return m.epgMap[channelID], nil
}

func (m *mockLiveTVRepo) Create(ctx context.Context, c *models.LiveTVChannel) (int64, error) {
	c.ID = len(m.channels) + 1
	m.channels = append(m.channels, *c)
	return int64(c.ID), nil
}

func (m *mockLiveTVRepo) Update(ctx context.Context, c *models.LiveTVChannel) error {
	for i, existing := range m.channels {
		if existing.ID == c.ID {
			m.channels[i] = *c
			return nil
		}
	}
	return nil
}

func (m *mockLiveTVRepo) Delete(ctx context.Context, id int) error {
	var remaining []models.LiveTVChannel
	for _, c := range m.channels {
		if c.ID != id {
			remaining = append(remaining, c)
		}
	}
	m.channels = remaining
	return nil
}

func (m *mockLiveTVRepo) SaveEPG(ctx context.Context, channelID int64, epg []models.EPG) error {
	if m.epgMap == nil {
		m.epgMap = make(map[int][]models.EPG)
	}
	m.epgMap[int(channelID)] = epg
	return nil
}

func (m *mockLiveTVRepo) UpdateYoutubeURL(ctx context.Context, id int, url string) error {
	for i, c := range m.channels {
		if c.ID == id {
			m.channels[i].YoutubeURL = &url
			return nil
		}
	}
	return nil
}

func TestLiveTVService_ListChannels(t *testing.T) {
	streamURL := "http://example.com/stream.m3u8"
	mock := &mockLiveTVRepo{
		channels: []models.LiveTVChannel{
			{ID: 1, Name: "Channel 1", StreamURL: &streamURL},
		},
		epgMap: map[int][]models.EPG{
			1: {
				{ID: 10, ChannelID: 1, ProgramTitle: "Show 1"},
			},
		},
	}

	service := NewLiveTVService(mock)
	channels, err := service.ListChannels(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}

	c := channels[0]
	if c.Name != "Channel 1" {
		t.Errorf("expected Channel 1, got %s", c.Name)
	}

	if len(c.EPG) != 1 || c.EPG[0].ProgramTitle != "Show 1" {
		t.Errorf("expected Show 1 in EPG, got %v", c.EPG)
	}
}

func TestLiveTVService_CreateChannel(t *testing.T) {
	mock := &mockLiveTVRepo{}
	service := NewLiveTVService(mock)

	newChan := &models.LiveTVChannel{Name: "New Channel"}
	id, err := service.CreateChannel(context.Background(), newChan)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id != 1 {
		t.Errorf("expected generated ID 1, got %d", id)
	}

	if len(mock.channels) != 1 || mock.channels[0].Name != "New Channel" {
		t.Errorf("expected channels list to contain 'New Channel', got %v", mock.channels)
	}
}

func TestLiveTVService_UpdateChannel(t *testing.T) {
	mock := &mockLiveTVRepo{
		channels: []models.LiveTVChannel{
			{ID: 1, Name: "Old Name"},
		},
	}
	service := NewLiveTVService(mock)

	updateChan := &models.LiveTVChannel{ID: 1, Name: "Updated Name"}
	err := service.UpdateChannel(context.Background(), updateChan)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock.channels[0].Name != "Updated Name" {
		t.Errorf("expected name to be 'Updated Name', got '%s'", mock.channels[0].Name)
	}
}

func TestLiveTVService_DeleteChannel(t *testing.T) {
	mock := &mockLiveTVRepo{
		channels: []models.LiveTVChannel{
			{ID: 1, Name: "Channel 1"},
			{ID: 2, Name: "Channel 2"},
		},
	}
	service := NewLiveTVService(mock)

	err := service.DeleteChannel(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.channels) != 1 || mock.channels[0].ID != 2 {
		t.Errorf("expected only Channel 2 to remain, got %v", mock.channels)
	}
}

func TestLiveTVService_SaveEPG(t *testing.T) {
	mock := &mockLiveTVRepo{}
	service := NewLiveTVService(mock)

	epgEntries := []models.EPG{
		{ProgramTitle: "Live Match"},
	}

	err := service.SaveEPG(context.Background(), 5, epgEntries)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.epgMap[5]) != 1 || mock.epgMap[5][0].ProgramTitle != "Live Match" {
		t.Errorf("expected EPG to contain 'Live Match', got %v", mock.epgMap[5])
	}
}

func TestLiveTVService_UpdateYoutubeURL(t *testing.T) {
	mock := &mockLiveTVRepo{
		channels: []models.LiveTVChannel{
			{ID: 1, Name: "Channel 1"},
		},
	}
	service := NewLiveTVService(mock)

	err := service.UpdateYoutubeURL(context.Background(), 1, "https://youtube.com/live")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock.channels[0].YoutubeURL == nil || *mock.channels[0].YoutubeURL != "https://youtube.com/live" {
		t.Errorf("expected Youtube URL update to succeed, got %v", mock.channels[0].YoutubeURL)
	}
}
