(function (global) {
    const createMobileSessionHandlers = (deps) => {
        const buildAuthPayload = () => {
            if (deps.authMode() === 'register') {
                return {
                    username: deps.form.username.trim(),
                    email: deps.form.email.trim(),
                    phone: deps.form.phone.trim(),
                    password: deps.form.password,
                    captchaId: deps.captchaId(),
                    captchaCode: deps.form.captchaCode.trim(),
                    smsCode: deps.form.smsCode.trim()
                };
            }

            if (deps.loginMethod() === 'sms') {
                return {
                    phone: deps.form.phone.trim(),
                    smsCode: deps.form.smsCode.trim()
                };
            }

            return {
                account: deps.form.account.trim(),
                password: deps.form.password,
                captchaId: deps.captchaId(),
                captchaCode: deps.form.captchaCode.trim()
            };
        };

        const applyAuthenticatedSession = ({ token, user, successMessage }) => {
            deps.setToken(token);
            localStorage.setItem('token', token);
            deps.setAuthenticated(true);
            deps.setUser(user);
            deps.syncProfileForm(user);
            deps.fetchOccupationOptions();
            if (successMessage) {
                deps.showToast(successMessage);
            }
            deps.startPolling();
            deps.reconcileRouteState({ replace: true });
        };

        const handleAuth = async () => {
            deps.setLoading(true);
            const endpoint = deps.authMode() === 'login' ? '/auth/login' : '/auth/register';
            const payload = buildAuthPayload();
            const res = await deps.request(endpoint, 'POST', payload);
            deps.setLoading(false);

            if (res) {
                if (deps.authMode() === 'register') {
                    deps.showToast('注册成功，请登录');
                    deps.setAuthMode('login');
                    deps.setLoginMethod('password');
                    deps.form.account = deps.form.email.trim();
                    deps.form.password = '';
                    deps.form.captchaCode = '';
                    deps.form.smsCode = '';
                    deps.fetchCaptcha();
                    return;
                }

                applyAuthenticatedSession({
                    token: res.token,
                    user: res.user,
                    successMessage: '登录成功'
                });
                return;
            }

            if (deps.requiresGraphCaptcha()) {
                deps.fetchCaptcha();
            }
        };

        const getUserInfo = async () => {
            const res = await deps.request('/user');
            if (res) {
                deps.setUser(res);
                deps.syncProfileForm(res);
                deps.fetchOccupationOptions();
                deps.setAuthenticated(true);
                deps.startPolling();
                deps.reconcileRouteState({ replace: true });
                return;
            }

            deps.setAuthenticated(false);
            deps.reconcileRouteState({ replace: true });
        };

        const logout = () => {
            deps.setToken('');
            localStorage.removeItem('token');
            deps.setAuthenticated(false);
            deps.setUser({});
            deps.syncProfileForm({});
            deps.stopPolling();
            deps.resetAlerts();
            deps.resetFamily();
            deps.resetChat();
            deps.reconcileRouteState({ replace: true });
        };

        return {
            buildAuthPayload,
            handleAuth,
            getUserInfo,
            logout
        };
    };

    global.SentinelSession = {
        createMobileSessionHandlers
    };
})(window);
