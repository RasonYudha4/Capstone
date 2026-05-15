package schemas

type AssessmentOption struct {
    ID          string `json:"id"`
    Code        string `json:"code"`
    Description string `json:"description"`
}

type StandardOption struct {
    ID          string             `json:"id"`
    Code        string             `json:"code"`
    Description string             `json:"description"`
    Assessments []AssessmentOption `json:"assessments"`
}

type ServiceOption struct {
    ID          string           `json:"id"`
    Code        string           `json:"code"`
    Description string           `json:"description"`
    Standards   []StandardOption `json:"standards"`
}

type FormOptionsResponse struct {
    Services      []ServiceOption      `json:"services"`
    DocumentTypes []DocumentTypeOption `json:"document_types"`
}

type DocumentTypeOption struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}