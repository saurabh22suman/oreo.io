# Sample Datasets for Oreo.io

This directory contains sample datasets for testing and demonstrating the data management capabilities of Oreo.io.

## 📊 Available Datasets

### 🛩️ Transportation (`transportation/`)

#### Airlines Flights Data (`airlines_flights_data.csv`)
- **Description**: Comprehensive flight booking data from various Indian airlines
- **Size**: ~24.9 MB, 300,153 rows
- **Columns**: 12 fields including airline, flight details, pricing, and booking information
- **Use Cases**: 
  - Data analysis and visualization testing
  - Performance benchmarking with large datasets
  - Time-series analysis (departure/arrival times)
  - Price analysis and forecasting
  - Route optimization studies

**Schema:**
```
index           - Unique record identifier
airline         - Airline name (SpiceJet, AirAsia, Vistara, etc.)
flight          - Flight number
source_city     - Departure city
departure_time  - Departure time category (Morning, Evening, etc.)
stops           - Number of stops (zero, one, etc.)
arrival_time    - Arrival time category
destination_city- Arrival city
class           - Seat class (Economy, Business)
duration        - Flight duration in hours
days_left       - Days until departure
price           - Ticket price in local currency
```

**Sample Data Preview:**
```
airline,flight,source_city,departure_time,stops,arrival_time,destination_city,class,duration,days_left,price
SpiceJet,SG-8709,Delhi,Evening,zero,Night,Mumbai,Economy,2.17,1,5953
AirAsia,I5-764,Delhi,Early_Morning,zero,Early_Morning,Mumbai,Economy,2.17,1,5956
Vistara,UK-995,Delhi,Morning,zero,Afternoon,Mumbai,Economy,2.25,1,5955
```

### 📁 Directory Structure

```
sample-data/
├── README.md              # This documentation
├── transportation/        # Transportation and logistics data
│   └── airlines_flights_data.csv
├── users/                 # User and customer data (planned)
├── finance/              # Financial and business data (planned)
└── mixed/                # Complex datasets for testing (planned)
```

## 🎯 Usage Examples

### 1. **Data Import Testing**
Use the airlines dataset to test data import functionality:
- CSV parsing and validation
- Large file handling (24.9 MB)
- Data type inference and conversion
- Error handling for malformed data

### 2. **Performance Benchmarking**
With 300K+ rows, this dataset is perfect for:
- Database query performance testing
- API response time optimization
- Frontend pagination and filtering
- Search functionality stress testing

### 3. **Analytics and Visualization**
Rich data for creating meaningful insights:
- Price distribution analysis
- Route popularity mapping
- Airline performance comparison
- Seasonal booking trends
- Duration vs. price correlation

### 4. **Business Intelligence**
Real-world scenarios for BI features:
- Revenue analysis by airline
- Popular routes identification
- Booking pattern analysis
- Price optimization strategies

## 🚀 Integration with Oreo.io

### Backend API Endpoints
```
GET    /api/v1/sample-data/                    # List all available datasets
GET    /api/v1/sample-data/transportation/     # List transportation datasets
GET    /api/v1/sample-data/transportation/airlines_flights_data/info  # Dataset metadata
GET    /api/v1/sample-data/transportation/airlines_flights_data/download  # Download CSV
POST   /api/v1/datasets/import/sample/         # Import sample dataset
```

### Frontend Features
- Dataset preview and exploration
- Interactive data visualization
- Schema definition and validation
- Data transformation and cleaning
- Export and sharing capabilities

## 📈 Dataset Statistics

| Metric | Value |
|--------|-------|
| **Total Records** | 300,153 |
| **File Size** | 24.9 MB |
| **Columns** | 12 |
| **Airlines** | 6 major carriers |
| **Routes** | Multiple city pairs |
| **Price Range** | Varies by route and class |
| **Data Quality** | Clean, structured data |

## 🛠️ Development Guidelines

### Adding New Datasets
1. Create appropriate category directory
2. Use descriptive filenames
3. Include UTF-8 encoded CSV with headers
4. Update this README with dataset information
5. Add sample queries and use cases

### File Naming Convention
- Use lowercase with underscores: `dataset_name.csv`
- Include version if applicable: `dataset_name_v2.csv`
- Use descriptive names: `airlines_flights_data.csv` not `data.csv`

### Data Quality Standards
- Include column headers in first row
- Use consistent date formats (ISO 8601 preferred)
- Handle missing values appropriately
- Validate data types and ranges
- Document any data transformations

## 🔧 Technical Notes

- All CSV files use UTF-8 encoding
- Date formats follow ISO 8601 where applicable
- Numeric values use standard decimal notation
- Text fields are properly escaped
- Files are optimized for both human and machine readability

## 📋 Future Additions

Planned datasets to be added:
- [ ] Customer demographics and behavior data
- [ ] Financial transactions and revenue data
- [ ] Product catalog and inventory data
- [ ] Marketing campaign performance data
- [ ] Operational metrics and KPIs
- [ ] Geographic and location-based data
