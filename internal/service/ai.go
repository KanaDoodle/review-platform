package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"

	"review-platform/config"
	"review-platform/internal/model"
	"review-platform/internal/repository"
)

type embeddedReview struct {
	Review    model.Review
	Embedding []float64
}

type AIService struct {
	cfg        config.AIConfig
	reviewRepo *repository.ReviewRepository

	mu      sync.RWMutex
	entries []embeddedReview

	cacheMu        sync.RWMutex
	embeddingCache map[string][]float64
}

func NewAIService(cfg config.AIConfig, reviewRepo *repository.ReviewRepository) *AIService {
	return &AIService{
		cfg:             cfg,
		reviewRepo:      reviewRepo,
		embeddingCache:  make(map[string][]float64),
	}
}

func (s *AIService) LoadReviewEmbeddings(ctx context.Context) error {
	reviews, err := s.reviewRepo.ListAll()
	if err != nil {
		return err
	}

	entries := make([]embeddedReview, 0, len(reviews))
	for _, review := range reviews {
		text := strings.TrimSpace(review.Content)
		if text == "" {
			continue
		}

		embedding, err := s.GetEmbedding(ctx, text)
		if err != nil {
			return err
		}

		entries = append(entries, embeddedReview{
			Review:    review,
			Embedding: embedding,
		})
	}

	s.mu.Lock()
	s.entries = entries
	s.mu.Unlock()

	return nil
}

func (s *AIService) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("empty text")
	}

	// ===== 1. 先查缓存 =====
	s.cacheMu.RLock()
	if emb, ok := s.embeddingCache[text]; ok {
		s.cacheMu.RUnlock()
		return emb, nil
	}
	s.cacheMu.RUnlock()

	// ===== 2. 调 API =====
	reqBody := map[string]interface{}{
		"input": text,
		"model": s.cfg.EmbeddingModel,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(s.cfg.BaseURL, "/") + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Message)
	}
	if len(result.Data) == 0 {
		return nil, errors.New("empty embedding response")
	}

	embedding := result.Data[0].Embedding

	// ===== 3. 写入缓存 =====
	s.cacheMu.Lock()
	s.embeddingCache[text] = embedding
	s.cacheMu.Unlock()

	return embedding, nil
}

type SearchResult struct {
	ReviewID int64   `json:"review_id"`
	ShopID   int64   `json:"shop_id"`
	UserID   int64   `json:"user_id"`
	Content  string  `json:"content"`
	Score    float64 `json:"score"`
}

func (s *AIService) SearchReviews(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("empty query")
	}

	queryEmbedding, err := s.GetEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	entries := make([]embeddedReview, len(s.entries))
	copy(entries, s.entries)
	s.mu.RUnlock()

	results := make([]SearchResult, 0, len(entries))
	for _, entry := range entries {
		embScore := cosineSimilarity(queryEmbedding, entry.Embedding)
		kwScore := keywordScore(query, entry.Review.Content)

		score := 0.7*embScore + 0.3*kwScore
		results = append(results, SearchResult{
			ReviewID: entry.Review.ID,
			ShopID:   entry.Review.ShopID,
			UserID:   entry.Review.UserID,
			Content:  entry.Review.Content,
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (s *AIService) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": s.cfg.ChatModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个点评平台助手，请基于提供的评论内容，用简洁、客观的中文总结商户评价，不要编造未提供的信息。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(s.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", errors.New(result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", errors.New("empty chat completion response")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func (s *AIService) SearchReviewsByShop(ctx context.Context, shopID int64, query string, topK int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("empty query")
	}

	queryEmbedding, err := s.GetEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	entries := make([]embeddedReview, len(s.entries))
	copy(entries, s.entries)
	s.mu.RUnlock()

	results := make([]SearchResult, 0)
	for _, entry := range entries {
		if entry.Review.ShopID != shopID {
			continue
		}
		embScore := cosineSimilarity(queryEmbedding, entry.Embedding)
		kwScore := keywordScore(query, entry.Review.Content)

		score := 0.7*embScore + 0.3*kwScore
		results = append(results, SearchResult{
			ReviewID: entry.Review.ID,
			ShopID:   entry.Review.ShopID,
			UserID:   entry.Review.UserID,
			Content:  entry.Review.Content,
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

type ShopSummaryResult struct {
	ShopID        int64          `json:"shop_id"`
	Query         string         `json:"query"`
	Summary       string         `json:"summary"`
	ReferenceDocs []SearchResult `json:"reference_docs"`
}

func (s *AIService) SummarizeShopReviews(ctx context.Context, shopID int64, query string, topK int) (*ShopSummaryResult, error) {
	results, err := s.SearchReviewsByShop(ctx, shopID, query, topK)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no relevant reviews found")
	}

	var sb strings.Builder
	sb.WriteString("请基于下面这些用户评论，回答问题并给出简洁总结。\n\n")
	sb.WriteString("用户问题：")
	sb.WriteString(query)
	sb.WriteString("\n\n")
	sb.WriteString("相关评论：\n")

	for i, item := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Content))
	}

	prompt := sb.String()

	summary, err := s.GenerateSummary(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &ShopSummaryResult{
		ShopID:        shopID,
		Query:         query,
		Summary:       summary,
		ReferenceDocs: results,
	}, nil
}

func keywordScore(query, content string) float64 {
	query = strings.ToLower(query)
	content = strings.ToLower(content)

	words := strings.Fields(query)
	if len(words) == 0 {
		return 0
	}

	hit := 0
	for _, w := range words {
		if strings.Contains(content, w) {
			hit++
		}
	}

	return float64(hit) / float64(len(words))
}
