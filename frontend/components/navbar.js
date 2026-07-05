function renderNavbar() {
    const navbarContainer = document.getElementById('navbar-container');
    if (!navbarContainer) return;

    const userJson = localStorage.getItem('current_user');
    const user = userJson ? JSON.parse(userJson) : null;
    const username = user ? user.username : '-';
    const fullName = user ? user.full_name : 'User';
    const initial = fullName.charAt(0).toUpperCase();

    navbarContainer.innerHTML = `
        <header class="app-header">
            <div class="header-left">
                <a href="/pages/dashboard.html" class="logo">
                    <i class="fa-solid fa-shield-halved"></i> GoPAMS
                </a>
            </div>
            <div class="header-right">
                <div class="notifications-bell">
                    <i class="fa-solid fa-bell"></i>
                    <span class="bell-badge">0</span>
                </div>
                <div class="header-user">
                    <div class="avatar">${initial}</div>
                    <span id="header-username">${username}</span>
                </div>
            </div>
        </header>
    `;
}
