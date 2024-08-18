# Reverse Captcha

## Overview

Reverse Captcha is an innovative web application that adds a unique twist to the traditional CAPTCHA system. Instead of asking users to prove they're human, it challenges them to demonstrate their ability to understand and respond to text descriptions with matching images. This project consists of a Go backend API and a frontend UI built with HTML, Tailwind CSS, and JavaScript.

![Demo of Reverse Captcha](assets/demo.png)

## Features

- User-friendly interface for requesting album information
- Artist search functionality
- Album selection from search results
- Reverse CAPTCHA challenge for accessing track lists

## How It Works

1. Users enter an artist name to retrieve a list of albums.
2. Upon selecting an album, users must complete a Reverse CAPTCHA challenge to access the track list.
3. The challenge presents a text description, and users must upload an image that matches the description.
4. If successful, the track list is displayed. If not, a new challenge is presented.

## Technical Stack

- Backend: Go
- Frontend: HTML, Tailwind CSS, JavaScript

## Installation

### Prerequisites

- Go
- OpenAI API Key (for Reverse CAPTCHA functionality)

### Backend Setup

1. Clone the repository:
    ```bash
    git clone https://github.com/danielshemesh/reversecaptcha
    cd reversecaptcha
    ```

2. Navigate to the backend directory:
    ```bash
    cd backend
    ```

3. Set the following environment variables in your terminal or system:

   - **Linux/MacOS:**
     ```bash
     export OPENAI_API_KEY=your_openai_api_key
     export OPENAI_API_BASE_URL=https://api.openai.com/v1
     ```

   - **Windows (Command Prompt):**
     ```cmd
     set OPENAI_API_KEY=your_openai_api_key
     set OPENAI_API_BASE_URL=https://api.openai.com/v1
     ```

   - **Windows (PowerShell):**
     ```powershell
     $env:OPENAI_API_KEY="your_openai_api_key"
     $env:OPENAI_API_BASE_URL="https://api.openai.com/v1"
     ```

   Replace `your_openai_api_key` with your actual OpenAI API key.
   If you're using a custom OpenAI API endpoint, replace the `OPENAI_BASE_URL` with the appropriate URL.

4. Install Go dependencies:
    ```bash
    go mod tidy
    ```

5. Run the Go server:
    ```bash
    go run backend.go
    ```

   The web app will be available at `http://localhost:8080`.
