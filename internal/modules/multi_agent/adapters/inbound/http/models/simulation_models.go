package models

type SimulationGeneratePackRequest struct {
	CaseType      string `json:"case_type"`
	TargetPersona string `json:"target_persona"`
	Difficulty    string `json:"difficulty"`
	Locale        string `json:"locale"`
}

type SimulationPackOption struct {
	Key        string `json:"key"`
	Text       string `json:"text"`
	RiskTag    string `json:"risk_tag"`
	ScoreDelta int    `json:"score_delta"`
	Rationale  string `json:"rationale"`
}

type SimulationPackStep struct {
	StepID         string                 `json:"step_id"`
	StepType       string                 `json:"step_type"`
	Narrative      string                 `json:"narrative"`
	Question       string                 `json:"question"`
	Options        []SimulationPackOption `json:"options"`
	KnowledgePoint string                 `json:"knowledge_point"`
	Difficulty     int                    `json:"difficulty"`
	TimeLimitSec   int                    `json:"time_limit_sec"`
}

type SimulationQuizPack struct {
	Title         string               `json:"title"`
	CaseType      string               `json:"case_type"`
	TargetPersona string               `json:"target_persona"`
	Difficulty    string               `json:"difficulty"`
	Intro         string               `json:"intro"`
	Steps         []SimulationPackStep `json:"steps"`
}

type SimulationGeneratePackResponse struct {
	PackID string             `json:"pack_id"`
	Pack   SimulationQuizPack `json:"pack"`
}

type SimulationStartSessionRequest struct {
	PackID string `json:"pack_id"`
}

type SimulationSubmitAnswerRequest struct {
	StepID    string `json:"step_id"`
	OptionKey string `json:"option_key"`
}

type SimulationSessionAnswer struct {
	StepID     string `json:"step_id"`
	StepType   string `json:"step_type"`
	OptionKey  string `json:"option_key"`
	OptionText string `json:"option_text"`
	RiskTag    string `json:"risk_tag"`
	ScoreDelta int    `json:"score_delta"`
	Rationale  string `json:"rationale"`
	AnsweredAt string `json:"answered_at"`
}

type SimulationSessionResult struct {
	TotalScore int      `json:"total_score"`
	Level      string   `json:"level"`
	Weaknesses []string `json:"weaknesses"`
	Strengths  []string `json:"strengths"`
	Advice     []string `json:"advice"`
}
