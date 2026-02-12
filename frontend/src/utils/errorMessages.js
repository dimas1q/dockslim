const LOCALE_STORAGE_KEY = 'dockslim_locale'

const RU_ERROR_MESSAGES = {
  unauthorized: 'Не авторизован',
  forbidden: 'Доступ запрещен',
  'invalid credentials': 'Неверные учетные данные',
  'invalid request body': 'Некорректное тело запроса',
  'request failed': 'Не удалось выполнить запрос',
  'csrf validation failed': 'Проверка CSRF не пройдена',
  'identifier is required': 'Нужно указать email или логин',
  'invalid email': 'Некорректный email',
  'invalid login': 'Некорректный логин',
  'invalid password': 'Некорректный пароль',
  'email already registered': 'Email уже зарегистрирован',
  'login already registered': 'Логин уже зарегистрирован',
  'failed to create user': 'Не удалось создать пользователя',
  'failed to login': 'Не удалось выполнить вход',
  'user not found': 'Пользователь не найден',
  'invalid token id': 'Некорректный идентификатор токена',
  'token name already exists': 'Токен с таким названием уже существует',
  'token not found': 'Токен не найден',
  'failed to create api token': 'Не удалось создать API-токен',
  'failed to list api tokens': 'Не удалось получить список API-токенов',
  'failed to revoke token': 'Не удалось отозвать токен',
  'no fields to update': 'Не указаны поля для обновления',
  'failed to update profile': 'Не удалось обновить профиль',
  'invalid user_id': 'Некорректный user_id',
  'invalid plan_id': 'Некорректный plan_id',
  'invalid valid_until format': 'Некорректный формат valid_until',
  'subscription service is not configured': 'Сервис подписок не настроен',
  'failed to fetch subscription': 'Не удалось получить подписку',
  'dashboard service is not configured': 'Сервис дашборда не настроен',
  'failed to fetch dashboard': 'Не удалось загрузить дашборд',
  'project not found': 'Проект не найден',
  'project with this name already exists': 'Проект с таким названием уже существует',
  'failed to create project': 'Не удалось создать проект',
  'failed to list projects': 'Не удалось получить список проектов',
  'failed to fetch project': 'Не удалось получить проект',
  'failed to fetch project role': 'Не удалось получить роль в проекте',
  'failed to update project': 'Не удалось обновить проект',
  'failed to delete project': 'Не удалось удалить проект',
  'invalid project id': 'Некорректный идентификатор проекта',
  'invalid budget id': 'Некорректный идентификатор бюджета',
  'invalid registry id': 'Некорректный идентификатор реестра',
  'invalid analysis id': 'Некорректный идентификатор анализа',
  'invalid from analysis id': 'Некорректный from analysis id',
  'invalid to analysis id': 'Некорректный to analysis id',
  'invalid from_analysis_id': 'Некорректный from_analysis_id',
  'invalid to_analysis_id': 'Некорректный to_analysis_id',
  'project_id is required': 'Поле project_id обязательно',
  'registry_id is required': 'Поле registry_id обязательно',
  'token does not match project': 'Токен не соответствует проекту',
  'invalid expires_at format': 'Некорректный формат expires_at',
  'invalid image': 'Некорректный образ',
  'invalid registry': 'Некорректный реестр',
  'registry not found': 'Реестр не найден',
  'registry with this name already exists': 'Реестр с таким названием уже существует',
  'failed to create registry': 'Не удалось создать реестр',
  'failed to list registries': 'Не удалось получить список реестров',
  'failed to update registry': 'Не удалось обновить реестр',
  'failed to delete registry': 'Не удалось удалить реестр',
  'failed to create analysis': 'Не удалось создать анализ',
  'failed to list analyses': 'Не удалось получить список анализов',
  'analysis not found': 'Анализ не найден',
  'failed to fetch analysis': 'Не удалось получить анализ',
  'analysis is running': 'Анализ уже выполняется',
  'failed to rerun analysis': 'Не удалось перезапустить анализ',
  'failed to delete analysis': 'Не удалось удалить анализ',
  'failed to compare analyses': 'Не удалось сравнить анализы',
  'analysis is not completed': 'Анализ еще не завершен',
  'both analyses must be completed': 'Оба анализа должны быть завершены',
  'analyses must be for the same image': 'Анализы должны быть для одного и того же образа',
  'feature not available on current plan': 'Функция недоступна на текущем тарифе',
  'failed to resolve feature access': 'Не удалось проверить доступ к функции',
  'failed to compare baseline': 'Не удалось выполнить baseline-сравнение',
  'baseline metrics unavailable': 'Метрики baseline недоступны',
  'no baseline analysis found': 'Базовый анализ не найден',
  'analysis must be completed': 'Анализ должен быть завершен',
  'baseline not found': 'Базовый анализ не найден',
  'failed to fetch history': 'Не удалось получить историю',
  'failed to fetch trends': 'Не удалось получить тренды',
  'failed to generate pdf export': 'Не удалось сформировать PDF-экспорт',
  'pdf export supports ascii characters only': 'PDF-экспорт поддерживает только ASCII-символы',
  'invalid status': 'Некорректный статус',
  'invalid from date': 'Некорректная дата начала',
  'invalid to date': 'Некорректная дата окончания',
  'invalid limit': 'Некорректный лимит',
  'invalid time': 'Некорректное время',
  'metric is required': 'Поле metric обязательно',
  'invalid metric': 'Некорректная метрика',
  'failed to fetch budgets': 'Не удалось получить бюджеты',
  'failed to save budget': 'Не удалось сохранить бюджет',
  'failed to create budget override': 'Не удалось создать override бюджета',
  'failed to update budget': 'Не удалось обновить бюджет',
  'failed to delete budget': 'Не удалось удалить бюджет',
  'budget not found': 'Бюджет не найден',
  'budget already exists': 'Бюджет уже существует',
  'invalid threshold': 'Некорректный порог',
  'invalid budget patch': 'Некорректные данные бюджета',
  'budget override for this image already exists': 'Override бюджета для этого образа уже существует',
  'failed to create ci token': 'Не удалось создать CI-токен',
  'failed to list ci tokens': 'Не удалось получить список CI-токенов',
  'ci token not found': 'CI-токен не найден',
  'ci token revoked': 'CI-токен отозван',
  'ci token expired': 'Срок действия CI-токена истек',
  'invalid ci token': 'Некорректный CI-токен',
  'ci token name already exists': 'CI-токен с таким названием уже существует',
  'invalid token name': 'Некорректное название токена',
  'invalid api token name': 'Некорректное название API-токена',
  'api token name already exists': 'API-токен с таким названием уже существует',
  'invalid api token': 'Некорректный API-токен',
  'api token revoked': 'API-токен отозван',
  'api token expired': 'Срок действия API-токена истек',
  'provider and repo are required': 'Поля provider и repo обязательны',
  'scm_token is required': 'Поле scm_token обязательно',
  'body_markdown is required': 'Поле body_markdown обязательно',
  'invalid project name': 'Некорректное имя проекта',
  'invalid project patch': 'Некорректные данные проекта',
  'project name already exists': 'Проект с таким названием уже существует',
  'user is not project owner': 'Недостаточно прав для операции',
  'invalid registry name': 'Некорректное имя реестра',
  'invalid registry url': 'Некорректный URL реестра',
  'invalid registry type': 'Некорректный тип реестра',
  'invalid registry patch': 'Некорректные данные реестра',
  'invalid registry credentials': 'Некорректные учетные данные реестра',
  'image registry does not match selected registry': 'Реестр образа не совпадает с выбранным реестром',
  'multiple registries match name': 'Найдено несколько реестров с таким именем',
  'missing registry identifier': 'Не указан идентификатор реестра',
  'failed to generate csrf token': 'Не удалось сгенерировать CSRF-токен',
  'missing authorization header': 'Отсутствует заголовок авторизации',
  'invalid authorization header': 'Некорректный заголовок авторизации',
  'empty bearer token': 'Пустой bearer-токен',
  'invalid token': 'Некорректный токен',
  'token missing kid header': 'В токене отсутствует заголовок kid',
  'unknown signing key': 'Неизвестный ключ подписи',
  'unexpected signing algorithm': 'Неподдерживаемый алгоритм подписи',
  'free plan comment limit is 2000 characters': 'Лимит комментария для Free-тарифа: 2000 символов',
  'failed to build compare report': 'Не удалось сформировать отчет сравнения',
}

const NETWORK_ERROR_PATTERNS = [
  'failed to fetch',
  'networkerror',
  'network error',
  'load failed',
]

export const getPreferredLocale = () => {
  if (typeof window === 'undefined' || !window.localStorage || typeof window.localStorage.getItem !== 'function') {
    return 'ru'
  }
  const stored = window.localStorage.getItem(LOCALE_STORAGE_KEY)
  if (stored === 'en' || stored === 'ru') {
    return stored
  }
  return 'ru'
}

const normalize = (message) => String(message || '').trim().toLowerCase()
const hasCyrillic = (value) => /[А-Яа-яЁё]/.test(value)
const hasLatin = (value) => /[A-Za-z]/.test(value)

const matchesNetworkError = (normalized) =>
  NETWORK_ERROR_PATTERNS.some((pattern) => normalized.includes(pattern))

export const localizeErrorMessage = (message, options = {}) => {
  const locale = options.locale || getPreferredLocale()
  const status = options.status

  const raw = String(message || '').trim()
  const normalized = normalize(raw)

  if (locale !== 'ru') {
    if (!raw) {
      return 'Request failed'
    }
    return raw
  }

  if (!raw) {
    if (status === 401) return 'Не авторизован'
    if (status === 403) return 'Доступ запрещен'
    if (status === 404) return 'Ресурс не найден'
    if (status === 409) return 'Конфликт данных'
    if (status != null && status >= 500) return 'Внутренняя ошибка сервера'
    return 'Не удалось выполнить запрос'
  }

  if (RU_ERROR_MESSAGES[normalized]) {
    return RU_ERROR_MESSAGES[normalized]
  }
  if (normalized.startsWith('github api returned status')) {
    return 'Ошибка GitHub API при публикации комментария'
  }
  if (normalized.startsWith('gitlab api returned status')) {
    return 'Ошибка GitLab API при публикации комментария'
  }
  if (matchesNetworkError(normalized)) {
    return 'Ошибка сети. Проверьте подключение и попробуйте снова'
  }

  if (status === 401) return 'Не авторизован'
  if (status === 403) return 'Доступ запрещен'
  if (status === 404) return 'Ресурс не найден'
  if (status === 409) return 'Конфликт данных'
  if (status != null && status >= 500) return 'Внутренняя ошибка сервера'

  if (!hasCyrillic(raw) && hasLatin(raw)) {
    return 'Не удалось выполнить запрос'
  }
  return raw
}
