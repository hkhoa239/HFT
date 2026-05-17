 Hệ thống này được thiết kế theo mô hình phân vai rõ ràng giữa:

* **Data Scientist (DS)** → tạo dữ liệu, train model, publish factor
* **Quant Researcher (QR)** → viết alpha strategy, chạy backtest, submit alpha
* **Portfolio Manager (PM)** → xem performance, correlation, đánh giá alpha/model

Ngoài ra còn có các lớp lưu trữ trung gian:

* Warehouse
* Published Data
* Model Repository
* Alpha Repository

Toàn bộ hệ thống xoay quanh:

* dữ liệu factor
* model ML
* alpha strategy
* metrics/performance

---

# 1. Tổng quan kiến trúc hệ thống

Hệ thống có 3 domain chính:

| Domain                | Vai trò           |
| --------------------- | ----------------- |
| Data Engineering / ML | Data Scientist    |
| Quant Research        | Quant Researcher  |
| Portfolio Analytics   | Portfolio Manager |

Các thành phần storage:

* Warehouse
* Published Data
* Model Repository
* Alpha Repository

Các loại dữ liệu chính:

* CSV lớn chứa factor data
* JSON metadata/model metrics
* Alpha performance JSON

---

# 2. FLOW: DATA SCIENTIST

Data Scientist có 3 workflow chính:

1. Analyst Script
2. Train Model
3. Publish Factor Data

---

# 2.1 Analyst Script Flow

Flow:

```text
Tạo Analyst Script
    ↓
Nhấn RUN
    ↓
Kết quả hiển thị ở terminal bên dưới khung script
```

Ý nghĩa:

* DS viết script phân tích exploratory
* Script chỉ để inspect/analyze data
* Không tạo artifact chính thức
* Output chỉ hiển thị realtime trên terminal/UI

Không có lưu DB/repository trong flow này.

---

# 2.2 Train Model Flow

Flow:

```text
Tạo Train Script
    ↓
Nhấn Train
    ↓
Kết quả hiển thị terminal
    ↓
Export?
```

Sau bước Export:

* model được lưu dưới dạng JSON metadata
* model artifact được lưu vào Model Repository

---

## 2.2.1 Model Repository

Model Repository chứa:

```text
Các JSON file
```

Mỗi JSON gồm:

| Field            | Ý nghĩa                |
| ---------------- | ---------------------- |
| pkl_path         | đường dẫn model pickle |
| train_date_start | ngày train bắt đầu     |
| train_date_end   | ngày train kết thúc    |
| eval_date_start  | ngày eval bắt đầu      |
| eval_date_end    | ngày eval kết thúc     |
| metrics          | metrics đánh giá model |

Ví dụ metrics:

```json
{
  "mse": 0.6,
  "precision": null,
  "recall": null
}
```

Ý nghĩa:

* Repository không chỉ lưu model binary
* Mà còn lưu metadata phục vụ reproducibility
* Metrics được dùng downstream cho PM review

---

# 2.3 Load / Evaluate Model Flow

Flow:

```text
Tải Model
    ↓
Kết quả:
1. Thực tế & dự đoán
2. Thống số metrics
```

Ý nghĩa:

* DS load model từ repository
* Chạy inference/evaluation
* So sánh prediction vs actual
* Sinh metrics đánh giá

Sau đó:

```text
Lưu metrics?
```

Nếu có:

* metrics được ghi lại vào JSON metadata trong Model Repository

---

# 3. FLOW: FACTOR DATA PIPELINE

Đây là pipeline dữ liệu quan trọng nhất hệ thống.

---

# 3.1 Tạo Factor Data

DS viết script:

```python
new_data = dataframe(...)
publish_data(new_data)
```

Flow:

```text
RUN
 ↓
Hàm publish_data kiểm tra data hợp lệ
 ↓
Xem data vừa tạo
 ↓
Publish data
```

---

# 3.2 Validation Layer

Function:

```text
publish_data()
```

chịu trách nhiệm:

* validate schema
* validate factor format
* validate compatibility với warehouse

Mục tiêu:

* tránh corrupt factor data
* đảm bảo merge được vào warehouse

---

# 3.3 Warehouse

Warehouse là:

```text
Một file CSV lớn
```

Schema:

| Column   |
| -------- |
| date     |
| factor_1 |
| factor_2 |
| ...      |

Ý nghĩa:

* warehouse là centralized factor storage
* mỗi factor là 1 column
* date là index chính

Flow:

```text
Factor mới
   ↓
Merge vào dataframe cũ
   ↓
Update warehouse CSV
```

---

# 3.4 Published Data

Sau publish:

```text
factor được lấy từ warehouse
 ↓
gộp vào dataframe chung
 ↓
Published Data
```

Published Data là:

* curated dataset
* stable dataset cho downstream consumer
* source cho Quant Researcher

Published Data cũng được lưu dạng:

```text
Một file CSV lớn
```

---

# 4. FLOW: QUANT RESEARCHER

QR chịu trách nhiệm:

* viết alpha
* chạy backtest
* submit strategy

---

# 4.1 Alpha Backtest Flow

Flow:

```text
Input:
- dataframe
- alpha script
```

Alpha script trả về:

```python
List[bool]
```

Ý nghĩa:

* strategy signal
* buy/sell condition
* mask/filter logic

---

Flow đầy đủ:

```text
Input dataframe
    +
Alpha script
    ↓
Run Backtest
    ↓
Kết quả Backtest
    ↓
Submit?
```

---

# 4.2 Alpha Repository

Nếu submit:

```text
Lưu code + performance alpha
```

vào:

```text
Alpha Repository
```

Repository lưu:

```text
JSON file
```

Schema:

| Field        | Meaning    |
| ------------ | ---------- |
| created_date | ngày tạo   |
| code         | alpha code |
| performance  | metrics    |

Ví dụ:

```json
{
  "PnL": 400,
  "Win_rate": 0.58
}
```

---

# 4.3 QR Analytics

QR có thể:

```text
Xem Performance Alpha đã submit
Xem Correlation Alpha đã submit
```

Điều này cho thấy:

* QR có visibility vào alpha history
* Có thể compare alpha
* Có thể tránh duplicated strategy

---

# 5. FLOW: PORTFOLIO MANAGER

PM là consumer cuối.

PM có thể:

| Feature               | Ý nghĩa              |
| --------------------- | -------------------- |
| Xem Model Performance | review ML quality    |
| Xem Correlation       | diversification/risk |
| Xem Performance/Code  | review alpha         |

---

# 5.1 PM đọc từ đâu?

PM lấy dữ liệu từ:

| Source           | Data                   |
| ---------------- | ---------------------- |
| Model Repository | model metrics          |
| Alpha Repository | alpha performance/code |

---

# 5.2 Correlation Analysis

PM dùng:

* correlation giữa alpha
* diversification analysis
* overlap detection

Mục tiêu:

* tránh alpha giống nhau
* giảm portfolio risk

---

# 5.3 Alpha Review

PM xem:

* alpha code
* performance metrics
* backtest quality

rồi quyết định:

* approve
* reject
* allocate capital

---

# 6. DATA STORAGE ARCHITECTURE

Hệ thống dùng hybrid storage:

| Storage                | Usage                |
| ---------------------- | -------------------- |
| CSV                    | factor warehouse     |
| JSON                   | metadata             |
| PKL                    | ML model             |
| Repository abstraction | logical organization |

---

# 6.1 Warehouse Pattern

Warehouse:

* append/update factor columns
* central source of truth
* dạng tabular timeseries

---

# 6.2 Repository Pattern

Có 2 repository:

| Repository       | Purpose         |
| ---------------- | --------------- |
| Model Repository | model lifecycle |
| Alpha Repository | alpha lifecycle |

---

# 7. END-TO-END SYSTEM FLOW

Toàn bộ pipeline:

```text
Data Scientist
    ↓
Generate factor
    ↓
Warehouse
    ↓
Published Data
    ↓
Quant Researcher
    ↓
Backtest Alpha
    ↓
Alpha Repository
    ↓
Portfolio Manager Review
```

Song song:

```text
Data Scientist
    ↓
Train Model
    ↓
Model Repository
    ↓
Portfolio Manager Review
```

---

# 8. KIẾN TRÚC LOGIC THỰC TẾ CỦA HỆ THỐNG

Đây là kiến trúc kiểu:

```text
Research Platform / Quant Platform
```

gần giống simplified:

* WorldQuant
* QuantConnect
* internal hedge fund tooling

Có các layer:

| Layer            | Responsibility          |
| ---------------- | ----------------------- |
| Data Layer       | Warehouse               |
| Research Layer   | Alpha scripts           |
| ML Layer         | Model training          |
| Analytics Layer  | Correlation/performance |
| Governance Layer | Submit/review flow      |

---

# 9. Ý NGHĨA CÁC MŨI TÊN NÉT ĐỨT

Các nét đứt thể hiện:

* data dependency
* read/write relation
* asynchronous logical flow

Ví dụ:

* metrics ghi vào repository
* PM đọc metrics
* QR submit alpha
* warehouse merge factor

---

# 10. TÓM TẮT VAI TRÒ

## Data Scientist

* tạo factor
* train model
* publish dataset
* evaluate metrics

## Quant Researcher

* viết alpha
* backtest
* submit strategy

## Portfolio Manager

* review performance
* correlation analysis
* evaluate alpha/model

---

# 11. CORE BUSINESS IDEA

Business flow thực sự của hệ thống là:

```text
Raw Data
 → Factor Engineering
 → Published Dataset
 → Alpha Research
 → Backtesting
 → Performance Evaluation
 → Portfolio Decision
```

Đây chính là toàn bộ business pipeline mà diagram đang mô tả.
