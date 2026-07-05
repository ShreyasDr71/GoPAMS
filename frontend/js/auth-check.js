(function() {
    const token = localStorage.getItem('jwt_token');
    const isLoginPage = window.location.pathname.includes('login.html');

    if (!token && !isLoginPage) {
        window.location.href = '/pages/login.html';
        return;
    }

    if (token && !isLoginPage) {
        // Pre-validate token against backend asynchronously
        fetch('/api/auth/me', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        }).then(res => {
            if (!res.ok) {
                if (res.status === 403) {
                    // Check if forced password change is required
                    res.json().then(data => {
                        if (data.error === 'password_change_required') {
                            window.location.href = '/pages/login.html?force_change=true';
                        } else {
                            localStorage.removeItem('jwt_token');
                            window.location.href = '/pages/login.html';
                        }
                    });
                } else {
                    localStorage.removeItem('jwt_token');
                    window.location.href = '/pages/login.html';
                }
            }
        }).catch(() => {
            // Network failure: let user remain in case of transient issues
        });
    }
})();
