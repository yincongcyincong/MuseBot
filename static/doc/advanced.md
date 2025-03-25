### 📝 **Custom Chain Command Documentation**

#### 📌 **Overview**
This document explains how to configure and use a custom chain command to execute a series of tasks, including HTTP requests and AI-based calculations, and finally generate investment advice.

---

## **📖 Configuration Fields**
| Field         | Type      | Description |
|---------------|-----------|-------------|
| `crontab`     | `string`  | Cron expression for scheduling the task, e.g., `"0 */1 * * * *"` triggers every minute |
| `command`     | `string`  | Task name, e.g., `"currency"` |
| `send_user`   | `string`  | The user who receives the result (optional) |
| `send_group`  | `string`  | The group that receives the result (optional) |
| `param`       | `object`  | Global parameters for the task, e.g., `currency_pair: "BTCUSDT"` |
| `chains`      | `array`   | An array of chained tasks |

### **📌 Chains**
The `chains` field contains multiple tasks grouped by `type`, executed sequentially.

#### **Task Types**
| Type        | Description                     |
|-------------|---------------------------------|
| `http`      | Executes an HTTP request        |
| `deepseek`  | Uses AI for processing and advice generation |

---

## **📖 Configuration Example**
Here’s an example JSON configuration:
```json
[
  {
    "crontab": "0 */1 * * * *",
    "command": "currency",
    "send_user": "",
    "send_group": "",
    "param": {
      "currency_pair": "BTCUSDT"
    },
    "chains": [
      {
        "type": "http",
        "tasks": [
          {
            "name": "task1",
            "http_param": {
              "url": "https://api.binance.com/api/v3/ticker/price?symbol={{.currency_pair}}",
              "method": "GET",
              "headers": {},
              "body": ""
            },
            "proxy": ""
          },
          {
            "name": "task2",
            "http_param": {
              "url": "https://api.binance.com/api/v3/ticker/price?symbol=DOGEUSDT",
              "method": "GET",
              "headers": {},
              "body": ""
            },
            "proxy": ""
          }
        ]
      },
      {
        "type": "deepseek",
        "tasks": [
          {
            "name": "task3",
            "template": "BTC price is {{.task1.price}}, Doge price is {{.task2.price}}, give me some advice about investment.",
            "proxy": ""
          }
        ]
      }
    ]
  }
]
```

---

## **📖 Execution Flow**
1. **Cron Trigger**
   - The task is triggered every minute as specified by the cron expression: `0 */1 * * * *`.
   - execute command `/currency`    
<img width="374" alt="aa92b3c9580da6926a48fc1fc5c37c03" src="https://github.com/user-attachments/assets/14415702-693a-4f8a-9403-71191d8649e2" />


2. **HTTP Task (`task1`)**
   - The system sends a request to:
     ```
     https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT
     ```
   - The API responds with the latest BTC price:
     ```json
     {
       "symbol": "BTCUSDT",
       "price": "45000.00"
     }
     ```
   - save data into `param` like：
     ```json
     {
      "task1": {
        "price": "45000.00", 
        "symbol": "BTCUSDT"
      }
     }
      ```

3. **AI Processing Task (`task2`)**
   - The AI task uses the response from `task1` and processes the template:
     ```
     BTC price is 45000.00, give me some advice about investment.
     ```
   - The AI generates investment advice based on the BTC price.

---

## **📖 Cron Expression Examples**
| Expression       | Description                  |
|------------------|------------------------------|
| `0 */1 * * * *`  | Runs every minute            |
| `0 0 * * * *`    | Runs every hour              |
| `0 0 9 * * *`    | Runs every day at 9 AM       |
| `0 0 0 1 * *`    | Runs on the first day of every month |


---

## **📖 Summary**
1. **Cron Trigger** → Executes every minute
2. **HTTP Request** → Fetches BTC price
3. **AI Analysis** → Generates investment advice
4. **Extensions** → Support multiple currencies and notifications

🚀 With this configuration, you can automate crypto price tracking and receive AI-generated investment advice regularly!

