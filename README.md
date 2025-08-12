# Oreo.io - Data Management Platform

A modern, role-based data management platform for collaborative dataset handling with schema enforcement, live editing, and automatic validation.

## 🎯 **Core Business Rules**

### **Data Flow Architecture:**

1. **Admin Dataset Upload & Schema Management**
   - Admin uploads target datasets (CSV, Excel, etc.)
   - Admin defines or edits data schemas for each dataset
   - System auto-infers schema if admin doesn't set one
   - Target datasets serve as the foundation for data operations

2. **User Data Contribution**
   - Users can add new data entries to existing target datasets
   - New data gets appended to the specified target dataset
   - All user contributions must conform to target schema

3. **Live Data Editing**
   - Users can edit their contributed data in real-time
   - Spreadsheet-style interface with immediate validation
   - Changes are validated against schema before saving

4. **Automatic Schema Drift Detection**
   - System continuously monitors for schema inconsistencies
   - Automatic notifications when data doesn't match expected schema
   - Alerts for potential data quality issues

---

### ✅ **Functional Requirements**

#### 1. **Authentication & Authorization**

* User registration & login
* Roles: Admin, Editor, Reviewer, Viewer
* Role-based access control per project

#### 2. **Project Management**

* Admin can create/edit/delete projects
* Admin can invite/add users to a project
* Permissions: read, write, edit with approval

#### 3. **Data Import & Target Management**

* **Admin Functions:**
  - Upload target datasets from desktop (CSV, Excel)
  - Connect to Google Drive or Google Sheets
  - Connect to SQL databases (e.g., MySQL, Postgres)
  - Set or edit schemas for target datasets
  - Auto-schema inference with manual override capability

#### 4. **Schema Management & Enforcement**

* **Automatic Schema Inference:**
  - Auto-detect data types, patterns, and constraints
  - Generate suggested schemas from uploaded data
  - Admin can review and modify inferred schemas

* **Schema Enforcement:**
  - Validate all new data against target schema
  - Real-time validation during data entry
  - Prevent invalid data submission

* **Schema Drift Detection:**
  - Monitor data consistency across entries
  - Automatic alerts for schema violations
  - Suggested schema updates when patterns change

#### 5. **Data Contribution & Live Editing**

* **User Data Entry:**
  - Add new rows to existing target datasets
  - Append data that conforms to target schema
  - Real-time validation feedback

* **Live Edit Interface:**
  - Spreadsheet-style grid view for data
  - Real-time validation and error highlighting
  - Auto-save with conflict resolution
  - Edit history and rollback capabilities

#### 6. **Business Rules & Validation**

* Define field-level rules (e.g., range, format, required)
* Custom validation logic per dataset
* Highlight invalid entries in the UI
* Rule sets reusable across similar datasets

#### 7. **Data Export**

* Save consolidated datasets to:
  * SQL database (via connection string)
  * Google Drive (as Sheet or Excel)
  * Local download (CSV/Excel)

#### 8. **User Dashboard**

* View assigned projects and accessible datasets
* Personal data contribution history
* Schema drift notifications and alerts
* Filter by last modified, status, etc.

---

### 🚀 **Non-Functional Requirements**

#### 1. **Performance**

* Load up to 100K rows smoothly in the editor
* Lazy loading/pagination for large datasets

#### 2. **Security**

* Secure OAuth for Google integration
* Encryption at rest and in transit
* Input sanitization and validation

#### 3. **Scalability**

* Microservice-friendly Go backend
* Deployable on Docker/Kubernetes
* Support multiple concurrent users per project

#### 4. **Usability**

* Clean, minimalist UI
* Mobile/tablet responsive layout
* Clear error messages and guidance

#### 5. **Reliability & Availability**

* Auto-save data periodically
* Retry mechanisms on failed exports
* Offline mode (optional)

#### 6. **Extensibility**

* Add more data sources (Dropbox, S3, etc.)
* Plugin support for custom validations or transformations

---

### 📦 Suggested Tech Stack

| Layer        | Technology                                    |
| ------------ | --------------------------------------------- |
| Backend API  | Golang (Gin/Fiber/Echo)                       |
| Frontend     | React.js or Svelte with TailwindCSS           |
| DB           | PostgreSQL or MySQL                           |
| File Storage | Google Drive API + Local filesystem           |
| Auth         | JWT + OAuth (Google)                          |
| Hosting      | Docker + Nginx + VPS or GCP                   |
| CI/CD        | GitHub Actions or GitLab CI                   |
| Optional     | Redis (caching), Kafka (data pipeline events) |

---

### ⚙️ Example Workflow

1. **Admin logs in** → Creates a project → Uploads dataset (CSV).
2. **Defines schema** → Sets business rules → Invites users.
3. **Users edit data** → Some edits require approval.
4. **Data validated & approved** → Exported to SQL DB or saved to Drive.

---

### 🚀 **Deployment Strategy**

#### **Environment Configuration**
- **Development**: Local Docker setup with hot reload
- **Staging**: staging.soloengine.in (Auto-deploy from `develop` branch)
- **Production**: prod.soloengine.in (Manual approval from `main` branch)
- **API**: api.soloengine.in (Versioned API endpoints)

#### **Infrastructure**
- **VPS**: 2 CPU, 2GB RAM, 30GB Storage
- **SSL**: Let's Encrypt wildcard (*.soloengine.in)
- **Database**: PostgreSQL with Redis caching
- **CI/CD**: GitHub Actions with strict deployment pipelines
- **Compliance**: GDPR-ready with audit logging

#### **Development Setup**
```bash
# Clone and setup
git clone https://github.com/saurabh22suman/oreo.io.git
cd oreo.io

# Start development environment
docker-compose up -d

# Access application
Frontend: http://localhost:3000
Backend: http://localhost:8080
Database: localhost:5432
```

#### **Testing & Quality Assurance**

##### **Integration Test Suite** ✅
We have a comprehensive, self-sufficient integration test suite that validates all API endpoints:

```bash
# Run all integration tests (self-contained, no setup required)
cd backend
go test -v ./tests/integration/ -timeout 5m

# Run specific test suites
go test -v ./tests/integration/ -run TestHealthEndpoints
go test -v ./tests/integration/ -run TestAuthenticationFlow
go test -v ./tests/integration/ -run TestProjectFlow
go test -v ./tests/integration/ -run TestSampleDataEndpoints
```

**Test Coverage:**
- ✅ **Health Endpoints** - Service health, database & Redis connectivity
- ✅ **Authentication Flow** - Registration, login, token refresh, error handling
- ✅ **Project Management** - Full CRUD operations with authorization
- ✅ **Sample Data Access** - File listing, info retrieval, preview functionality

**Key Features:**
- **Zero Dependencies** - No manual setup or configuration required
- **Self-Contained** - Creates unique test data, cleans up automatically
- **Fast Execution** - Complete test suite runs in ~3 seconds
- **Comprehensive** - Tests both success and error scenarios
- **CI/CD Ready** - Perfect for automated pipelines

##### **Test Results Summary**
```
✅ TestHealthEndpoints          - 3 test cases
✅ TestAuthenticationFlow       - 7 test cases  
✅ TestProjectFlow             - 6 test cases
✅ TestSampleDataEndpoints     - 4 test cases

Total: 20 test cases, Runtime: ~3.2s, Status: ALL PASSING
```

#### **Testing Strategy**
- **Integration Testing** ✅ Complete API endpoint validation with self-sufficient test suite
- **TDD Approach**: Write tests first, then implementation
- **Coverage**: Minimum 80% test coverage across all components
- **E2E Testing**: Cypress for critical user flows (planned)
- **Performance**: 100K rows handling validation
- **Security Testing**: Authentication, authorization, and data validation
- **Self-Contained Tests**: No external dependencies or manual setup required
