package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/yamlforge/yamlforge/internal/parser"
)

func formatFieldName(fieldName string) string {
	if fieldName == "" {
		return ""
	}
	
	// If it contains underscores, replace with spaces and title case each word
	if strings.Contains(fieldName, "_") {
		result := strings.ReplaceAll(fieldName, "_", " ")
		words := strings.Fields(result)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
			}
		}
		return strings.Join(words, " ")
	}
	
	// Special case for all uppercase: make it title case
	if strings.ToUpper(fieldName) == fieldName {
		return strings.ToUpper(string(fieldName[0])) + strings.ToLower(fieldName[1:])
	}
	
	// Otherwise, just capitalize the first letter and preserve the rest
	return strings.ToUpper(string(fieldName[0])) + fieldName[1:]
}

func GetHomeHTML(config *parser.Config, schema *parser.Schema, modelPermissions map[string]bool) string {
	modelsMenu := ""
	for modelName := range schema.Models {
		modelsMenu += fmt.Sprintf(`<li><a href="/%s">%s</a></li>`, strings.ToLower(modelName), modelName)
	}
	
	if config.Server.Auth.Type != "none" {
		modelsMenu += `<li style="margin-top: auto;"><a href="/logout" style="color: #e53e3e;">Logout</a></li>`
	}

	modelCards := ""
	for modelName := range schema.Models {
		canWrite := true
		if modelPermissions != nil {
			if hasPermission, exists := modelPermissions[modelName]; exists {
				canWrite = hasPermission
			}
		}
		
		addNewButton := ""
		if canWrite {
			addNewButton = fmt.Sprintf(`<a href="/%s/new" class="btn btn-secondary">Add New</a>`, strings.ToLower(modelName))
		}
		
		modelCards += fmt.Sprintf(`
		<div class="stat-card">
			<h3>%s</h3>
			<div class="stat-actions">
				<a href="/%s" class="btn btn-primary">View All</a>
				%s
			</div>
		</div>`, modelName, strings.ToLower(modelName), addNewButton)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Dashboard</title>
    <style>%s</style>
</head>
<body>
    <div class="app-container layout-sidebar">
        <nav class="sidebar">
            <div class="logo">
                <h1>%s</h1>
            </div>
            <ul class="nav-menu">
                <li><a href="/">Dashboard</a></li>
                %s
            </ul>
        </nav>
        <main class="main-content">
            <div class="page-header">
                <h2>Dashboard</h2>
            </div>
            <div class="dashboard">
                <div class="stats-grid">
                    %s
                </div>
            </div>
        </main>
    </div>
    <script>%s</script>
</body>
</html>`, config.App.Name, getCSS(), config.App.Name, modelsMenu, modelCards, getJS())
}

func GetListHTML(config *parser.Config, schema *parser.Schema, modelName string, model *parser.Model, canWrite bool) string {
	modelsMenu := ""
	for mName := range schema.Models {
		activeClass := ""
		if strings.ToLower(mName) == strings.ToLower(modelName) {
			activeClass = ` class="active"`
		}
		modelsMenu += fmt.Sprintf(`<li><a href="/%s"%s>%s</a></li>`, strings.ToLower(mName), activeClass, mName)
	}
	
	if config.Server.Auth.Type != "none" {
		modelsMenu += `<li style="margin-top: auto;"><a href="/logout" style="color: #e53e3e;">Logout</a></li>`
	}

	addNewButton := ""
	if canWrite {
		addNewButton = fmt.Sprintf(`<a href="/%s/new" class="btn btn-primary">Add New</a>`, strings.ToLower(modelName))
	}
	
	columnHeaders := ""
	columns := model.UI.List.Columns
	if len(columns) == 0 {
		for _, field := range model.Fields {
			if !field.AutoNow && !field.AutoNowAdd && field.Name != "id" {
				columns = append(columns, field.Name)
				if len(columns) >= 4 {
					break
				}
			}
		}
	}

	for _, col := range columns {
		columnHeaders += fmt.Sprintf("<th>%s</th>", formatFieldName(col))
	}

	columnsJSON, _ := json.Marshal(columns)
	searchableJSON, _ := json.Marshal(model.UI.List.Searchable)
	sortableJSON, _ := json.Marshal(model.UI.List.Sortable)

	modelInfo := buildModelInfoJSON(model)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - %s</title>
    <style>%s</style>
</head>
<body>
    <div class="app-container layout-sidebar">
        <nav class="sidebar">
            <div class="logo">
                <h1>%s</h1>
            </div>
            <ul class="nav-menu">
                <li><a href="/">Dashboard</a></li>
                %s
            </ul>
        </nav>
        <main class="main-content">
            <div class="page-header">
                <h2>%s</h2>
                <div class="page-actions">
                    %s
                </div>
            </div>
            <div class="list-container">
                <div class="search-bar">
                    <input type="text" id="search" placeholder="Search..." class="form-control">
                </div>
                <div class="data-table-container">
                    <table class="data-table" id="dataTable">
                        <thead>
                            <tr>
                                %s
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody id="tableBody">
                            <tr><td colspan="%d" class="loading">Loading...</td></tr>
                        </tbody>
                    </table>
                </div>
                <div class="pagination" id="pagination"></div>
            </div>
        </main>
    </div>
    <script>%s
    const modelName = '%s';
    const columns = %s;
    const searchable = %s;
    const sortable = %s;
    const modelInfo = %s;
    const canWrite = %t;

    document.addEventListener('DOMContentLoaded', () => {
        loadList(modelName, columns, searchable, sortable);
    });
    </script>
</body>
</html>`, modelName, config.App.Name, getCSS(), config.App.Name, modelsMenu, modelName, 
		addNewButton, columnHeaders, len(columns)+1, getJS(), 
		strings.ToLower(modelName), string(columnsJSON), string(searchableJSON), 
		string(sortableJSON), modelInfo, canWrite)
}

func GetFormHTML(config *parser.Config, schema *parser.Schema, modelName string, model *parser.Model, action string, recordId string, recordJSON string) string {
	isEdit := action == "edit"
	pageTitle := fmt.Sprintf("%s %s", "New", modelName)
	pageHeader := fmt.Sprintf("%s %s", "New", modelName)
	submitText := "Create"
	if isEdit {
		pageTitle = fmt.Sprintf("Edit %s %s", modelName, recordId)
		pageHeader = fmt.Sprintf("Edit %s %s", modelName, recordId)
		submitText = "Update"
	}

	modelsMenu := ""
	for mName := range schema.Models {
		activeClass := ""
		if strings.ToLower(mName) == strings.ToLower(modelName) {
			activeClass = ` class="active"`
		}
		modelsMenu += fmt.Sprintf(`<li><a href="/%s"%s>%s</a></li>`, strings.ToLower(mName), activeClass, mName)
	}
	
	if config.Server.Auth.Type != "none" {
		modelsMenu += `<li style="margin-top: auto;"><a href="/logout" style="color: #e53e3e;">Logout</a></li>`
	}

	formFields := ""
	formFieldNames := model.UI.Form.Fields
	if len(formFieldNames) == 0 {
		for _, field := range model.Fields {
			if !field.AutoNow && !field.AutoNowAdd && field.Name != "id" {
				formFieldNames = append(formFieldNames, field.Name)
			}
		}
	}

	for _, fieldName := range formFieldNames {
		var field *parser.Field
		for i := range model.Fields {
			if model.Fields[i].Name == fieldName {
				field = &model.Fields[i]
				break
			}
		}
		if field == nil {
			continue
		}

		formFields += generateFormField(field)
	}

	modelInfo := buildModelInfoJSON(model)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - %s</title>
    <style>%s</style>
</head>
<body>
    <div class="app-container layout-sidebar">
        <nav class="sidebar">
            <div class="logo">
                <h1>%s</h1>
            </div>
            <ul class="nav-menu">
                <li><a href="/">Dashboard</a></li>
                %s
            </ul>
        </nav>
        <main class="main-content">
            <div class="page-header">
                <h2>%s</h2>
            </div>
            <div class="form-container">
                <form id="modelForm" class="model-form">
                    %s
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">%s</button>
                        <a href="/%s" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </main>
    </div>
    <script>%s
    const modelName = '%s';
    const action = '%s';
    const recordData = %s;
    const recordId = recordData ? recordData.id : null;
    const modelInfo = %s;

    document.addEventListener('DOMContentLoaded', () => {
        if (action === 'edit' && recordData) {
            Object.keys(recordData).forEach(key => {
                const elem = document.getElementById(key);
                if (elem) {
                    if (elem.type === 'password') {
                        elem.placeholder = '(unchanged if empty)';
                    } else if (elem.type === 'checkbox') {
                        elem.checked = recordData[key];
                    } else if (elem.type === 'date' && recordData[key]) {
                        const dateValue = recordData[key].includes('T')
                            ? recordData[key].split('T')[0]
                            : recordData[key];
                        elem.value = dateValue;
                    } else {
                        elem.value = recordData[key] || '';
                    }
                }
            });
        } else if (action === 'create') {
            const formElements = document.querySelectorAll('[data-default]');
            formElements.forEach(elem => {
                const defaultValue = elem.getAttribute('data-default');
                if (defaultValue !== null && defaultValue !== '') {
                    if (elem.type === 'checkbox') {
                        elem.checked = (defaultValue === 'true' || defaultValue === '1');
                    } else if (elem.type === 'number') {
                        elem.value = parseFloat(defaultValue) || 0;
                    } else if (elem.tagName === 'SELECT') {
                        elem.value = defaultValue;
                    } else if (elem.tagName === 'TEXTAREA') {
                        elem.value = defaultValue;
                    } else {
                        elem.value = defaultValue;
                    }
                }
            });
        }

        handleForm(modelName, action, recordId);
    });
    </script>
</body>
</html>`, pageTitle, config.App.Name, getCSS(), config.App.Name, modelsMenu,
		pageHeader, formFields, submitText, strings.ToLower(modelName),
		getJS(), strings.ToLower(modelName), action, recordJSON, modelInfo)
}

func GetViewHTML(config *parser.Config, schema *parser.Schema, modelName string, model *parser.Model, recordId string, recordJSON string) string {
	modelsMenu := ""
	for mName := range schema.Models {
		activeClass := ""
		if strings.ToLower(mName) == strings.ToLower(modelName) {
			activeClass = ` class="active"`
		}
		modelsMenu += fmt.Sprintf(`<li><a href="/%s"%s>%s</a></li>`, strings.ToLower(mName), activeClass, mName)
	}
	
	if config.Server.Auth.Type != "none" {
		modelsMenu += `<li style="margin-top: auto;"><a href="/logout" style="color: #e53e3e;">Logout</a></li>`
	}

	modelInfo := buildModelInfoJSON(model)

	fieldDisplayLogic := ""
	for _, field := range model.Fields {
		fieldDisplayLogic += fmt.Sprintf(`
                    if (record.%s !== undefined && record.%s !== null && record.%s !== '') {
                        html += '<div class="detail-row"><div class="detail-label">%s</div><div class="detail-value">' + escapeHtml(formatDetailValue('%s', record.%s, modelInfo)) + '</div></div>';
                    }`, field.Name, field.Name, field.Name, formatFieldName(field.Name), field.Name, field.Name)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s Details - %s</title>
    <style>%s</style>
</head>
<body>
    <div class="app-container layout-sidebar">
        <nav class="sidebar">
            <div class="logo">
                <h1>%s</h1>
            </div>
            <ul class="nav-menu">
                <li><a href="/">Dashboard</a></li>
                %s
            </ul>
        </nav>
        <main class="main-content">
            <div class="page-header">
                <h2>%s Details</h2>
                <div class="page-actions">
                    <a href="/%s/%s/edit" class="btn btn-primary">Edit</a>
                    <button onclick="deleteRecord('%s', '%s')" class="btn btn-danger">Delete</button>
                    <a href="/%s" class="btn btn-secondary">Back to List</a>
                </div>
            </div>
            <div class="detail-container">
                <div class="detail-card" id="detailCard">
                    <div class="loading">Loading...</div>
                </div>
            </div>
        </main>
    </div>
    <script>%s
    const recordId = '%s';
    const modelInfo = %s;

    function formatDetailValue(fieldName, value, modelInfo) {
        if (modelInfo.fields[fieldName] && modelInfo.fields[fieldName].type === 'password') {
            return '********';
        }
        
        if (modelInfo.fields[fieldName] && modelInfo.fields[fieldName].options) {
            const options = modelInfo.fields[fieldName].options;
            if (options[value] !== undefined) {
                return options[value];
            }
        }
        
        if (typeof value === 'boolean') {
            return value ? 'Yes' : 'No';
        }
        
        if (fieldName.includes('date') || fieldName.includes('_at')) {
            if (!value) return '';
            const date = new Date(value);
            if (isNaN(date.getTime())) return value;
            return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
        }
        
        return value || '';
    }

    document.addEventListener('DOMContentLoaded', () => {
        const record = %s;
        if (record) {
            const detailCard = document.getElementById('detailCard');
            let html = '';
            %s
            detailCard.innerHTML = html || '<div class="empty-state">No data available</div>';
        } else {
            document.getElementById('detailCard').innerHTML = '<div class="error-message">Failed to load record</div>';
        }
    });

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
    </script>
</body>
</html>`, modelName, config.App.Name, getCSS(), config.App.Name, modelsMenu, modelName,
		strings.ToLower(modelName), recordId, strings.ToLower(modelName), recordId,
		strings.ToLower(modelName), getJS(), recordId, modelInfo, recordJSON, fieldDisplayLogic)
}

func generateFormField(field *parser.Field) string {
	required := ""
	if field.Required {
		required = " required"
	}

	switch field.Type {
	case parser.FieldTypeText, parser.FieldTypeEmail, parser.FieldTypePhone, parser.FieldTypeURL, parser.FieldTypePassword:
		inputType := "text"
		if field.Type == parser.FieldTypeEmail {
			inputType = "email"
		} else if field.Type == parser.FieldTypePhone {
			inputType = "tel"
		} else if field.Type == parser.FieldTypeURL {
			inputType = "url"
		} else if field.Type == parser.FieldTypePassword {
			inputType = "password"
		}
		
		minAttr := ""
		maxAttr := ""
		if field.Min != nil {
			minAttr = fmt.Sprintf(` minlength="%d"`, *field.Min)
		}
		if field.Max != nil {
			maxAttr = fmt.Sprintf(` maxlength="%d"`, *field.Max)
		}
		
		defaultValue := ""
		if field.Default != nil && field.Type != parser.FieldTypePassword {
			defaultValue = fmt.Sprintf(` data-default="%v"`, field.Default)
		}
		
		return fmt.Sprintf(`<div class="form-group">
        <label for="%s">%s%s</label>
        <input type="%s" id="%s" name="%s" class="form-control"%s%s%s%s>
    </div>`, field.Name, formatFieldName(field.Name), requiredStar(field.Required), inputType, field.Name, field.Name, required, minAttr, maxAttr, defaultValue)
	
	case parser.FieldTypeMarkdown, parser.FieldTypeJSON:
		defaultValue := ""
		if field.Default != nil {
			defaultValue = fmt.Sprintf(` data-default="%v"`, field.Default)
		}
		
		return fmt.Sprintf(`<div class="form-group">
        <label for="%s">%s%s</label>
        <textarea id="%s" name="%s" class="form-control" rows="5"%s%s></textarea>
    </div>`, field.Name, formatFieldName(field.Name), requiredStar(field.Required), field.Name, field.Name, required, defaultValue)
	
	case parser.FieldTypeBoolean:
		defaultValue := ""
		if field.Default != nil {
			defaultValue = fmt.Sprintf(` data-default="%v"`, field.Default)
		}
		
		return fmt.Sprintf(`<div class="form-group">
        <label for="%s">
            <input type="checkbox" id="%s" name="%s"%s>
            %s
        </label>
    </div>`, field.Name, field.Name, field.Name, defaultValue, formatFieldName(field.Name))
	
	case parser.FieldTypeNumber:
		minAttr := ""
		maxAttr := ""
		if field.Min != nil {
			minAttr = fmt.Sprintf(` min="%d"`, *field.Min)
		}
		if field.Max != nil {
			maxAttr = fmt.Sprintf(` max="%d"`, *field.Max)
		}
		defaultVal := ""
		if field.Default != nil {
			defaultVal = fmt.Sprintf(` data-default="%v"`, field.Default)
		}
		
		return fmt.Sprintf(`<div class="form-group">
        <label for="%s">%s%s</label>
        <input type="number" id="%s" name="%s" class="form-control"%s%s%s%s>
    </div>`, field.Name, formatFieldName(field.Name), requiredStar(field.Required), field.Name, field.Name, required, minAttr, maxAttr, defaultVal)
	
	case parser.FieldTypeEnum:
		options := ""
		defaultAttr := ""
		for _, opt := range field.Options {
			displayName := strings.ReplaceAll(strings.Title(strings.ReplaceAll(opt, "_", " ")), " ", " ")
			options += fmt.Sprintf(`<option value="%s">%s</option>`, opt, displayName)
		}
		
		if field.Default != nil {
			defaultAttr = fmt.Sprintf(` data-default="%v"`, field.Default)
		}
		
		return fmt.Sprintf(`<div class="form-group">
        <label for="%s">%s%s</label>
        <select id="%s" name="%s" class="form-control"%s%s>
            %s
        </select>
    </div>`, field.Name, formatFieldName(field.Name), requiredStar(field.Required), field.Name, field.Name, required, defaultAttr, options)
	
	case parser.FieldTypeDate, parser.FieldTypeDatetime, parser.FieldTypeTime:
		inputType := "date"
		if field.Type == parser.FieldTypeDatetime {
			inputType = "datetime-local"
		} else if field.Type == parser.FieldTypeTime {
			inputType = "time"
		}
		
		return fmt.Sprintf(`<div class="form-group">
        <label for="%s">%s%s</label>
        <input type="%s" id="%s" name="%s" class="form-control"%s>
    </div>`, field.Name, formatFieldName(field.Name), requiredStar(field.Required), inputType, field.Name, field.Name, required)
	
	default:
		return fmt.Sprintf(`<div class="form-group">
        <label for="%s">%s%s</label>
        <input type="text" id="%s" name="%s" class="form-control"%s>
    </div>`, field.Name, formatFieldName(field.Name), requiredStar(field.Required), field.Name, field.Name, required)
	}
}

func requiredStar(required bool) string {
	if required {
		return "*"
	}
	return ""
}

func buildModelInfoJSON(model *parser.Model) string {
	fieldInfo := make(map[string]map[string]any)
	for _, field := range model.Fields {
		info := make(map[string]any)
		info["type"] = string(field.Type)
		
		if field.Type == parser.FieldTypeEnum && len(field.Options) > 0 {
			options := make(map[string]string)
			for _, opt := range field.Options {
				displayName := strings.ReplaceAll(strings.Title(strings.ReplaceAll(opt, "_", " ")), " ", " ")
				options[opt] = displayName
			}
			info["options"] = options
		}
		
		fieldInfo[field.Name] = info
	}
	
	modelInfo := map[string]any{
		"fields": fieldInfo,
	}
	
	jsonBytes, _ := json.Marshal(modelInfo)
	return string(jsonBytes)
}

func GetLoginHTML(config *parser.Config) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Login</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
        }
        .login-container {
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            position: relative;
            overflow: hidden;
        }
        .login-container::before {
            content: '';
            position: absolute;
            width: 200%%;
            height: 200%%;
            background: radial-gradient(circle, rgba(255,255,255,0.1) 1px, transparent 1px);
            background-size: 50px 50px;
            animation: backgroundMove 60s linear infinite;
        }
        @keyframes backgroundMove {
            0%% { transform: translate(0, 0); }
            100%% { transform: translate(50px, 50px); }
        }
        .login-box {
            background: rgba(255, 255, 255, 0.98);
            backdrop-filter: blur(10px);
            padding: 3rem;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3), 0 0 0 1px rgba(255,255,255,0.1);
            width: 100%%;
            max-width: 420px;
            position: relative;
            z-index: 1;
            animation: slideUp 0.5s ease-out;
        }
        @keyframes slideUp {
            from {
                opacity: 0;
                transform: translateY(30px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        .login-header {
            text-align: center;
            margin-bottom: 2.5rem;
        }
        .login-header h1 {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 0.5rem;
            letter-spacing: -0.5px;
        }
        .login-header p {
            color: #64748b;
            font-size: 0.95rem;
            font-weight: 400;
        }
        .login-form .form-group {
            margin-bottom: 1.5rem;
        }
        .login-form label {
            display: block;
            margin-bottom: 0.5rem;
            color: #4a5568;
            font-weight: 500;
        }
        .login-form input {
            width: 100%%;
            padding: 0.875rem 1rem;
            border: 2px solid #e2e8f0;
            border-radius: 10px;
            font-size: 1rem;
            transition: all 0.3s;
            background: #f8fafc;
        }
        .login-form input:focus {
            outline: none;
            border-color: #667eea;
            background: white;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }
        .login-btn {
            width: 100%%;
            padding: 0.875rem;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            border: none;
            border-radius: 10px;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
            position: relative;
            overflow: hidden;
        }
        .login-btn::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 100%%;
            height: 100%%;
            background: linear-gradient(135deg, #764ba2 0%%, #667eea 100%%);
            opacity: 0;
            transition: opacity 0.3s;
        }
        .login-btn:hover::before {
            opacity: 1;
        }
        .login-btn span {
            position: relative;
            z-index: 1;
        }
        .login-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }
        .login-btn:active {
            transform: translateY(0);
        }
        .login-btn:disabled {
            background: #cbd5e0;
            cursor: not-allowed;
            transform: none;
        }
        .error-message {
            background: linear-gradient(135deg, #fee2e2 0%%, #fecaca 100%%);
            color: #dc2626;
            padding: 1rem;
            border-radius: 10px;
            margin-bottom: 1.5rem;
            display: none;
            border-left: 4px solid #dc2626;
            animation: shake 0.5s ease-in-out;
        }
        @keyframes shake {
            0%%, 100%% { transform: translateX(0); }
            25%% { transform: translateX(-10px); }
            75%% { transform: translateX(10px); }
        }
        .success-message {
            background: linear-gradient(135deg, #dcfce7 0%%, #bbf7d0 100%%);
            color: #16a34a;
            padding: 1rem;
            border-radius: 10px;
            margin-bottom: 1.5rem;
            display: none;
            border-left: 4px solid #16a34a;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="login-box">
            <div class="login-header">
                <h1>%s</h1>
                <p>Sign in to your account</p>
            </div>
            
            <div id="errorMessage" class="error-message"></div>
            <div id="successMessage" class="success-message"></div>
            
            <form id="loginForm" class="login-form">
                <div class="form-group">
                    <label for="username">Username or Email</label>
                    <input type="text" id="username" name="username" required autofocus>
                </div>
                
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>
                
                <button type="submit" class="login-btn" id="loginBtn"><span>Sign In</span></button>
            </form>
            
        </div>
    </div>
    
    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const errorMsg = document.getElementById('errorMessage');
            const successMsg = document.getElementById('successMessage');
            const loginBtn = document.getElementById('loginBtn');
            
            errorMsg.style.display = 'none';
            successMsg.style.display = 'none';
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            
            loginBtn.disabled = true;
            loginBtn.textContent = 'Signing in...';
            
            try {
                const response = await fetch('/api/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ username, password }),
                });
                
                const data = await response.json();
                
                if (response.ok && data.success) {
                    successMsg.textContent = 'Login successful! Redirecting...';
                    successMsg.style.display = 'block';
                    
                    const returnUrl = new URLSearchParams(window.location.search).get('return') || '/';
                    setTimeout(() => {
                        window.location.href = returnUrl;
                    }, 1000);
                } else {
                    errorMsg.textContent = data.error || 'Invalid credentials';
                    errorMsg.style.display = 'block';
                    loginBtn.disabled = false;
                    loginBtn.textContent = 'Sign In';
                }
            } catch (error) {
                errorMsg.textContent = 'An error occurred. Please try again.';
                errorMsg.style.display = 'block';
                loginBtn.disabled = false;
                loginBtn.textContent = 'Sign In';
            }
        });
    </script>
</body>
</html>`, config.App.Name, config.App.Name)
}