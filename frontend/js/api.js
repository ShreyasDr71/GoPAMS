const API_BASE = ''; // Same host

function getAuthToken() {
    return localStorage.getItem('jwt_token') || '';
}

function setAuthToken(token) {
    if (token) {
        localStorage.setItem('jwt_token', token);
    } else {
        localStorage.removeItem('jwt_token');
    }
}

async function apiFetch(endpoint, method = 'GET', body = null) {
    const token = getAuthToken();
    const options = {
        method: method,
        headers: {
            'Accept': 'application/json'
        }
    };

    if (token) {
        options.headers['Authorization'] = `Bearer ${token}`;
    }

    if (body) {
        options.headers['Content-Type'] = 'application/json';
        options.body = JSON.stringify(body);
    }

    const res = await fetch(endpoint, options);
    const data = await res.json();

    if (!res.ok) {
        // Handle token expiration/revocation
        if (res.status === 401 && !window.location.pathname.includes('login.html')) {
            setAuthToken('');
            window.location.href = '/pages/login.html?session_expired=true';
            return;
        }
        throw new Error(data.message || 'API request failed');
    }

    return data;
}

function escapeHtml(str) {
    if (!str) return '';
    return str
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}
