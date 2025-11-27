// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

const successHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Successful - Magistrala</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: #073764;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 48px;
            max-width: 500px;
            width: 100%;
            text-align: center;
            animation: slideIn 0.4s ease-out;
        }
        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(-20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        .logo-container {
            margin-bottom: 32px;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .logo {
            width: 100%;
            max-width: 400px;
            height: auto;
            animation: logoFadeIn 0.8s ease-out 0.3s both;
        }
        @keyframes logoFadeIn {
            from {
                opacity: 0;
                transform: scale(0.8);
            }
            to {
                opacity: 1;
                transform: scale(1);
            }
        }
        .success-icon {
            width: 60px;
            height: 60px;
            margin: 0 auto 24px;
            background: #073764;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: scaleIn 0.5s ease-out 0.5s both;
            position: relative;
        }
        @keyframes scaleIn {
            from {
                transform: scale(0);
            }
            to {
                transform: scale(1);
            }
        }
        .checkmark {
            width: 50px;
            height: 50px;
        }
        .checkmark-path {
            stroke: white;
            stroke-width: 4;
            fill: none;
            stroke-linecap: round;
            stroke-linejoin: round;
            stroke-dasharray: 100;
            stroke-dashoffset: 100;
            animation: drawCheck 0.6s ease-out 0.8s forwards;
        }
        @keyframes drawCheck {
            to {
                stroke-dashoffset: 0;
            }
        }
        h1 {
            color: #073764;
            font-size: 32px;
            font-weight: 600;
            margin-bottom: 16px;
        }
        p {
            color: #4a5568;
            font-size: 18px;
            line-height: 1.6;
            margin-bottom: 12px;
        }
        .footer {
            margin-top: 32px;
            padding-top: 24px;
            border-top: 1px solid #e2e8f0;
            color: #a0aec0;
            font-size: 14px;
        }
        @media (max-width: 768px) {
            .container {
                padding: 40px 32px;
            }
            .logo {
                max-width: 350px;
            }
            h1 {
                font-size: 28px;
            }
            p {
                font-size: 17px;
            }
        }
        @media (max-width: 600px) {
            .container {
                padding: 32px 24px;
            }
            .logo {
                max-width: 300px;
            }
            h1 {
                font-size: 26px;
            }
            p {
                font-size: 16px;
            }
            .success-icon,
            .error-icon {
                width: 55px;
                height: 55px;
            }
            .checkmark {
                width: 40px;
                height: 40px;
            }
        }
        @media (max-width: 480px) {
            .container {
                padding: 28px 20px;
            }
            .logo {
                max-width: 260px;
            }
            h1 {
                font-size: 24px;
            }
            p {
                font-size: 15px;
            }
            .success-icon,
            .error-icon {
                width: 50px;
                height: 50px;
            }
            .checkmark {
                width: 35px;
                height: 35px;
            }
        }
        @media (max-width: 360px) {
            .container {
                padding: 24px 16px;
            }
            .logo {
                max-width: 220px;
            }
            h1 {
                font-size: 22px;
            }
            p {
                font-size: 14px;
            }
            .success-icon,
            .error-icon {
                width: 45px;
                height: 45px;
            }
            .checkmark {
                width: 32px;
                height: 32px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo-container">
            <img src="https://cloud.magistrala.absmach.eu/_next/static/media/Magistrala_logo_landscape_white.59ea595a.svg"
                 alt="Magistrala Logo"
                 class="logo"
                 style="filter: brightness(0) saturate(100%) invert(18%) sepia(58%) saturate(1976%) hue-rotate(189deg) brightness(96%) contrast(103%);">
        </div>

        <div class="success-icon">
            <svg class="checkmark" viewBox="0 0 52 52">
                <path class="checkmark-path" d="M14 27l10 10 18-20"/>
            </svg>
        </div>

        <h1>Authentication Successful!</h1>
        <p>You have been successfully authenticated.</p>
        <p>You can now close this window and return to the CLI.</p>

        <div class="footer">
            Powered by SuperMQ
        </div>
    </div>
</body>
</html>`

const errorHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Failed - Magistrala</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: #073764;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 48px;
            max-width: 500px;
            width: 100%;
            text-align: center;
            animation: slideIn 0.4s ease-out;
        }
        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(-20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        .logo-container {
            margin-bottom: 32px;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .logo {
            width: 100%;
            max-width: 400px;
            height: auto;
            animation: logoFadeIn 0.8s ease-out 0.3s both;
        }
        @keyframes logoFadeIn {
            from {
                opacity: 0;
                transform: scale(0.8);
            }
            to {
                opacity: 1;
                transform: scale(1);
            }
        }
        .error-icon {
            width: 60px;
            height: 60px;
            margin: 0 auto 24px;
            background: #dc2626;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: shake 0.5s ease-out 0.5s both;
        }
        @keyframes shake {
            0%, 100% { transform: translateX(0); }
            25% { transform: translateX(-10px); }
            75% { transform: translateX(10px); }
        }
        .cross {
            width: 38px;
            height: 38px;
            position: relative;
        }
        .cross::before,
        .cross::after {
            content: '';
            position: absolute;
            width: 3px;
            height: 38px;
            background: white;
            left: 50%;
            top: 0;
            border-radius: 2px;
        }
        .cross::before {
            transform: translateX(-50%) rotate(45deg);
        }
        .cross::after {
            transform: translateX(-50%) rotate(-45deg);
        }
        h1 {
            color: #073764;
            font-size: 32px;
            font-weight: 600;
            margin-bottom: 16px;
        }
        p {
            color: #4a5568;
            font-size: 18px;
            line-height: 1.6;
            margin-bottom: 12px;
        }
        .error-message {
            background: #fef2f2;
            border: 1px solid #fecaca;
            border-radius: 8px;
            padding: 16px;
            margin: 24px 0;
            color: #991b1b;
            font-family: monospace;
            font-size: 14px;
            word-break: break-word;
        }
        .footer {
            margin-top: 32px;
            padding-top: 24px;
            border-top: 1px solid #e2e8f0;
            color: #a0aec0;
            font-size: 14px;
        }
        @media (max-width: 768px) {
            .container {
                padding: 40px 32px;
            }
            .logo {
                max-width: 350px;
            }
            h1 {
                font-size: 28px;
            }
            p {
                font-size: 17px;
            }
        }
        @media (max-width: 600px) {
            .container {
                padding: 32px 24px;
            }
            .logo {
                max-width: 300px;
            }
            h1 {
                font-size: 26px;
            }
            p {
                font-size: 16px;
            }
            .success-icon,
            .error-icon {
                width: 70px;
                height: 70px;
            }
            .cross {
                width: 28px;
                height: 28px;
            }
            .cross::before,
            .cross::after {
                height: 28px;
            }
        }
        @media (max-width: 480px) {
            .container {
                padding: 28px 20px;
            }
            .logo {
                max-width: 260px;
            }
            h1 {
                font-size: 24px;
            }
            p {
                font-size: 15px;
            }
            .success-icon,
            .error-icon {
                width: 50px;
                height: 50px;
            }
            .cross {
                width: 32px;
                height: 32px;
            }
            .cross::before,
            .cross::after {
                height: 32px;
            }
        }
        @media (max-width: 360px) {
            .container {
                padding: 24px 16px;
            }
            .logo {
                max-width: 220px;
            }
            h1 {
                font-size: 22px;
            }
            p {
                font-size: 14px;
            }
            .success-icon,
            .error-icon {
                width: 45px;
                height: 45px;
            }
            .cross {
                width: 28px;
                height: 28px;
            }
            .cross::before,
            .cross::after {
                height: 28px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo-container">
            <img src="https://cloud.magistrala.absmach.eu/_next/static/media/Magistrala_logo_landscape_white.59ea595a.svg"
                 alt="Magistrala Logo"
                 class="logo"
                 style="filter: brightness(0) saturate(100%) invert(18%) sepia(58%) saturate(1976%) hue-rotate(189deg) brightness(96%) contrast(103%);">
        </div>

        <div class="error-icon">
            <div class="cross"></div>
        </div>

        <h1>Authentication Failed</h1>
        <p>We encountered an error during authentication.</p>

        <div class="error-message">
            {{ERROR_MESSAGE}}
        </div>

        <p>Please close this window and try again.</p>

        <div class="footer">
            Powered by SuperMQ
        </div>
    </div>
</body>
</html>`
