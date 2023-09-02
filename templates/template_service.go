package templates

import (
	"database/sql"
	"fmt"
	"unimock/util"
)

const InsertQuery = "INSERT INTO templates (name, body, subsystem) VALUES (?,?,?)"
const SelectAllQuery = "SELECT * FROM templates"
const UpdateQuery = "UPDATE templates SET name = ?, body = ?, subsystem = ? where id = ?"
const DeleteQuery = "DELETE FROM templates WHERE id = ?"

type TemplateService struct {
	templates map[int64]*Template
	db        *sql.DB
}

func NewService(db *sql.DB) *TemplateService {
	return &TemplateService{
		templates: make(map[int64]*Template),
		db:        db,
	}
}

func (service *TemplateService) GetTemplates() []*Template {
	templateValues := make([]*Template, 0, len(service.templates))

	for _, value := range service.templates {
		templateValues = append(templateValues, value)
	}

	return templateValues
}

func (service *TemplateService) GetTemplatesWithoutBody() []Template {
	templateValues := make([]Template, 0, len(service.templates))

	for _, value := range service.templates {
		templateValues = append(templateValues, Template{
			Id:        value.Id,
			Name:      value.Name,
			Subsystem: value.Subsystem,
		})
	}

	return templateValues
}

func (service *TemplateService) GetTemplateById(id int64) (*Template, error) {
	template, ok := service.templates[id]

	if !ok {
		return nil, &TemplateNotFoundException{
			message: fmt.Sprintf("Шаблон с id = %d не найден", id),
		}
	} else {
		return template, nil
	}
}

func (service *TemplateService) AddTemplate(template *Template) error {
	if !template.validate() {
		return &TemplateValidationException{message: "Не указано имя шаблона"}
	}
	insertStatement, err := service.db.Prepare(InsertQuery)
	if err != nil {
		return err
	}
	res, err := insertStatement.Exec(template.Name, template.Body, template.Subsystem)
	if err != nil {
		return err
	}

	template.Id, err = res.LastInsertId()
	if err != nil {
		return err
	}

	err = insertStatement.Close()
	if err != nil {
		return err
	}

	service.templates[template.Id] = template
	return nil
}

func (service *TemplateService) UpdateTemplate(template *Template) error {
	if !template.validate() {
		return &TemplateValidationException{message: "Не указано имя шаблона"}
	}
	updateStatement, err := service.db.Prepare(UpdateQuery)
	if err != nil {
		return err
	}
	_, err = updateStatement.Exec(template.Name, template.Body, template.Subsystem, template.Id)
	if err != nil {
		return err
	}

	err = updateStatement.Close()
	if err != nil {
		return err
	}

	service.templates[template.Id] = template
	return nil
}

func (service *TemplateService) DeleteTemplate(id int64) error {
	deleteStatement, err := service.db.Prepare(DeleteQuery)
	if err != nil {
		return err
	}
	_, err = deleteStatement.Exec(id)
	if err != nil {
		return err
	}

	err = deleteStatement.Close()
	if err != nil {
		return err
	}

	delete(service.templates, id)
	return nil
}

func (service *TemplateService) UpdateFromDb() error {
	rows, err := service.db.Query(SelectAllQuery)
	if err != nil {
		return err
	}

	service.templates = make(map[int64]*Template)

	for rows.Next() {
		var t Template
		err = rows.Scan(&t.Id, &t.Name, &t.Body, &t.Subsystem)
		if err != nil {
			return err
		}

		service.templates[t.Id] = &t
	}
	return nil
}

func (service *TemplateService) ProcessMessage(templateId int64, message *util.Message) (*util.Message, error) {
	template, err := service.GetTemplateById(templateId)
	if err != nil {
		return nil, err
	}
	return template.ProcessMessage(message), nil
}
