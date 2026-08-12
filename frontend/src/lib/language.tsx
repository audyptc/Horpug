import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

export type Language = 'en' | 'th'

const translations = {
  en: {
    brandSubtitle: 'Studio Dashboard',
    searchPlaceholder: 'Search videos, analytics...',
    openMenu: 'Open menu',
    collapseSidebar: 'Collapse sidebar',
    expandSidebar: 'Expand sidebar',
    switchToLight: 'Switch to light mode',
    switchToDark: 'Switch to dark mode',
    notifications: 'Notifications',
    language: 'Language',
    accountMenu: 'Account menu',
    myAccount: 'My Account',
    profile: 'Profile',
    channelSettings: 'Channel settings',
    signOut: 'Sign out',
    mobileMenuTitle: 'Menu',
    mobileMenuDescription: 'Navigate to your admin sections',
    adminMenuLabel: 'Admin menu',
    mainMenu: 'Main Menu',
    dashboard: 'Dashboard',
    dashboardOverview: 'Overview of your channel performance and latest content actions.',
    totalViews: 'Total Views',
    totalViewsDetail: '+12.4% this month',
    subscribers: 'Subscribers',
    subscribersDetail: '+1,203 new',
    watchTime: 'Watch Time',
    watchTimeDetail: '+8.7% this week',
    recentUploadQueue: 'Recent Upload Queue',
    recentUploadQueueDescription: 'Video processing status from your latest uploads.',
    tableVideo: 'Video',
    tableStatus: 'Status',
    tableReach: 'Reach',
    tableLastUpdate: 'Last Update',
    statusPublished: 'Published',
    statusProcessing: 'Processing',
    statusDraft: 'Draft',
    reachPending: 'Pending',
    reachNotPublished: 'Not Published',
    timeHoursAgo: '2 hours ago',
    timeMinutesAgo: '12 minutes ago',
    timeYesterday: 'Yesterday',
    menuUsers: 'Users',
    menuUsersDescription: 'Manage user accounts and their assigned roles.',
    menuRoles: 'Roles',
    menuRolesDescription: 'Manage roles and the menu permissions granted to them.',
    menuPermissions: 'Permissions',
    menuPermissionsDescription: 'Browse the catalog of permission actions.',
    menuDormitories: 'Dormitories',
    menuDormitoriesDescription: 'Manage dormitories and their assigned managers.',
    menuActivityLogs: 'Activity Logs',
    menuActivityLogsDescription: 'Review recorded system activity.',
    liveApiData: 'Live API Data',
    recordsLoaded: 'records loaded',
    loading: 'Loading...',
    noRecords: 'No records found.',
    resourceLoadError: 'Failed to load data from the API.',
    statusActive: 'Active',
    statusInactive: 'Inactive',
    loginTitle: 'Sign in',
    loginSubtitle: 'Sign in with your Horpug admin account.',
    username: 'Username',
    password: 'Password',
    signIn: 'Sign in',
    signingIn: 'Signing in...',
    loginError: 'Sign-in failed. Please try again.',
  },
  th: {
    brandSubtitle: 'แดชบอร์ดสตูดิโอ',
    searchPlaceholder: 'ค้นหาวิดีโอ, ข้อมูลวิเคราะห์...',
    openMenu: 'เปิดเมนู',
    collapseSidebar: 'ย่อแถบเมนู',
    expandSidebar: 'ขยายแถบเมนู',
    switchToLight: 'สลับเป็นโหมดสว่าง',
    switchToDark: 'สลับเป็นโหมดมืด',
    notifications: 'การแจ้งเตือน',
    language: 'ภาษา',
    accountMenu: 'เมนูบัญชี',
    myAccount: 'บัญชีของฉัน',
    profile: 'โปรไฟล์',
    channelSettings: 'การตั้งค่าช่อง',
    signOut: 'ออกจากระบบ',
    mobileMenuTitle: 'เมนู',
    mobileMenuDescription: 'ไปยังส่วนต่าง ๆ ของระบบแอดมิน',
    adminMenuLabel: 'เมนูแอดมิน',
    mainMenu: 'เมนูหลัก',
    dashboard: 'แดชบอร์ด',
    dashboardOverview: 'ภาพรวมประสิทธิภาพช่องและกิจกรรมล่าสุดของเนื้อหา',
    totalViews: 'ยอดวิวทั้งหมด',
    totalViewsDetail: '+12.4% เดือนนี้',
    subscribers: 'ผู้ติดตาม',
    subscribersDetail: '+1,203 คนใหม่',
    watchTime: 'เวลาการรับชม',
    watchTimeDetail: '+8.7% สัปดาห์นี้',
    recentUploadQueue: 'คิวอัปโหลดล่าสุด',
    recentUploadQueueDescription: 'สถานะการประมวลผลวิดีโอจากการอัปโหลดล่าสุดของคุณ',
    tableVideo: 'วิดีโอ',
    tableStatus: 'สถานะ',
    tableReach: 'การเข้าถึง',
    tableLastUpdate: 'อัปเดตล่าสุด',
    statusPublished: 'เผยแพร่แล้ว',
    statusProcessing: 'กำลังประมวลผล',
    statusDraft: 'ฉบับร่าง',
    reachPending: 'รอดำเนินการ',
    reachNotPublished: 'ยังไม่เผยแพร่',
    timeHoursAgo: '2 ชั่วโมงที่แล้ว',
    timeMinutesAgo: '12 นาทีที่แล้ว',
    timeYesterday: 'เมื่อวาน',
    menuUsers: 'ผู้ใช้งาน',
    menuUsersDescription: 'จัดการบัญชีผู้ใช้และบทบาทที่ได้รับมอบหมาย',
    menuRoles: 'บทบาท',
    menuRolesDescription: 'จัดการบทบาทและสิทธิ์การเข้าถึงเมนูต่าง ๆ',
    menuPermissions: 'สิทธิ์การใช้งาน',
    menuPermissionsDescription: 'ดูรายการสิทธิ์การใช้งานทั้งหมดในระบบ',
    menuDormitories: 'หอพัก',
    menuDormitoriesDescription: 'จัดการหอพักและผู้ดูแลที่ได้รับมอบหมาย',
    menuActivityLogs: 'บันทึกกิจกรรม',
    menuActivityLogsDescription: 'ตรวจสอบบันทึกกิจกรรมที่เกิดขึ้นในระบบ',
    liveApiData: 'ข้อมูลจาก API แบบเรียลไทม์',
    recordsLoaded: 'รายการที่โหลดแล้ว',
    loading: 'กำลังโหลด...',
    noRecords: 'ไม่พบข้อมูล',
    resourceLoadError: 'โหลดข้อมูลจาก API ไม่สำเร็จ',
    statusActive: 'ใช้งาน',
    statusInactive: 'ปิดใช้งาน',
    loginTitle: 'เข้าสู่ระบบ',
    loginSubtitle: 'เข้าสู่ระบบด้วยบัญชีแอดมิน Horpug ของคุณ',
    username: 'ชื่อผู้ใช้',
    password: 'รหัสผ่าน',
    signIn: 'เข้าสู่ระบบ',
    signingIn: 'กำลังเข้าสู่ระบบ...',
    loginError: 'เข้าสู่ระบบไม่สำเร็จ กรุณาลองใหม่อีกครั้ง',
  },
} as const

export type TranslationKey = keyof (typeof translations)['en']

type LanguageContextValue = {
  language: Language
  setLanguage: (language: Language) => void
  t: (key: TranslationKey) => string
}

const LanguageContext = createContext<LanguageContextValue | null>(null)

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<Language>(
    () => (localStorage.getItem('language') as Language | null) ?? 'en'
  )

  useEffect(() => {
    localStorage.setItem('language', language)
    document.documentElement.lang = language
  }, [language])

  const value = useMemo<LanguageContextValue>(
    () => ({
      language,
      setLanguage,
      t: (key) => translations[language][key],
    }),
    [language]
  )

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>
}

export function useLanguage() {
  const context = useContext(LanguageContext)
  if (!context) throw new Error('useLanguage must be used within LanguageProvider')
  return context
}
