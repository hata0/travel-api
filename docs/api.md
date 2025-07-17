# Travel API 仕様書

## 概要

旅行プランを管理するためのWeb APIです。

## エンドポイント

### 1. 旅行プラン一覧取得

- **Method:** `GET`
- **Path:** `/trips`
- **Description:** 登録されている旅行プランの一覧を取得します。
- **URL Parameters:** なし
- **Request Body:** なし
- **Success Response:**
    - **Code:** `200 OK`
    - **Content:**
      ```json
      {
        "trips": [
          {
            "id": "string",
            "name": "string",
            "created_at": "string (RFC3339)",
            "updated_at": "string (RFC3339)"
          }
        ]
      }
      ```
- **Error Response:**
    - **Code:** `500 Internal Server Error`
    - **Content:**
      ```json
      {
        "message": "internal server error"
      }
      ```

### 2. 旅行プラン取得

- **Method:** `GET`
- **Path:** `/trips/:trip_id`
- **Description:** 指定されたIDの旅行プランを1件取得します。
- **URL Parameters:**
    - `trip_id` (string, required): 旅行プランID
- **Request Body:** なし
- **Success Response:**
    - **Code:** `200 OK`
    - **Content:**
      ```json
      {
        "trip": {
          "id": "string",
          "name": "string",
          "created_at": "string (RFC3339)",
          "updated_at": "string (RFC3339)"
        }
      }
      ```
- **Error Response:**
    - **Code:** `404 Not Found`
    - **Content:**
      ```json
      {
        "message": "trip not found"
      }
      ```
    - **Code:** `500 Internal Server Error`
    - **Content:**
      ```json
      {
        "message": "internal server error"
      }
      ```

### 3. 旅行プラン作成

- **Method:** `POST`
- **Path:** `/trips`
- **Description:** 新しい旅行プランを作成します。
- **URL Parameters:** なし
- **Request Body:**
  ```json
  {
    "name": "string (required)"
  }
  ```
- **Success Response:**
    - **Code:** `200 OK`
    - **Content:**
      ```json
      {
        "message": "success"
      }
      ```
- **Error Response:**
    - **Code:** `500 Internal Server Error`
    - **Content:**
      ```json
      {
        "message": "internal server error"
      }
      ```

### 4. 旅行プラン更新

- **Method:** `PUT`
- **Path:** `/trips/:trip_id`
- **Description:** 指定されたIDの旅行プランを更新します。
- **URL Parameters:**
    - `trip_id` (string, required): 旅行プランID
- **Request Body:**
  ```json
  {
    "name": "string (required)"
  }
  ```
- **Success Response:**
    - **Code:** `200 OK`
    - **Content:**
      ```json
      {
        "message": "success"
      }
      ```
- **Error Response:**
    - **Code:** `404 Not Found`
    - **Content:**
      ```json
      {
        "message": "trip not found"
      }
      ```
    - **Code:** `500 Internal Server Error`
    - **Content:**
      ```json
      {
        "message": "internal server error"
      }
      ```

### 5. 旅行プラン削除

- **Method:** `DELETE`
- **Path:** `/trips/:trip_id`
- **Description:** 指定されたIDの旅行プランを削除します。
- **URL Parameters:**
    - `trip_id` (string, required): 旅行プランID
- **Request Body:** なし
- **Success Response:**
    - **Code:** `200 OK`
    - **Content:**
      ```json
      {
        "message": "success"
      }
      ```
- **Error Response:**
    - **Code:** `404 Not Found`
    - **Content:**
      ```json
      {
        "message": "trip not found"
      }
      ```
    - **Code:** `500 Internal Server Error`
    - **Content:**
      ```json
      {
        "message": "internal server error"
      }
      ```
