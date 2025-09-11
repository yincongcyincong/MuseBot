import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

i18n
    .use(LanguageDetector) // 自动检测语言
    .use(initReactI18next) // React 绑定
    .init({
        resources: {
            en: {
                translation: {
                    dashboard_name: "Bot Dashboard",
                    bot_choose: "Select Bot"
                }
            },
            zh: {
                translation: {
                    dashboard_name: "机器人仪表板",
                    bot_choose: "选取机器人"
                }
            }
        },
        fallbackLng: 'en',
        interpolation: {
            escapeValue: false,
        }
    });

export default i18n;
