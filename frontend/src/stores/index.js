/**
 * Stores Index
 * =============
 * Export semua stores dari satu tempat.
 */

// Auth store
export {
    currentUser,
    authToken,
    isAuthenticated,
    isAdmin,
    isLoading,
    setAuth,
    clearAuth,
    setLoading,
} from "./auth.js";

// Site Settings store
export {
    siteSettings,
    siteSettingsLoading,
    appName,
    appDescription,
    aboutText,
    visi,
    misi,
    footerText,
    logoFullUrl,
    contactInfo,
    refreshSiteSettings,
    initSiteSettings,
} from "./siteSettings.js";
