// @ts-check
;(() => {
  /**
   * @type {import('/@/types/theme').Theme}
   */
  const lightTheme = {
    version: 2,
    basic: {
      accent: {
        primary: {
          default: '#006B8F',
          background: '#006B8F',
          inactive: 'rgba(0, 107, 143, 0.55)',
          fallback: '#006B8F'
        },
        notification: {
          default: '#FF6A00',
          background: '#FF6A00',
          fallback: '#FF6A00'
        },
        online: '#00C853',
        error: '#FF1744',
        focus: '#42B8C8B8'
      },
      background: {
        primary: {
          default: '#FFFFFF',
          border: '#0B0D12',
          fallback: '#FFFFFF'
        },
        secondary: {
          default: '#EEF1F5',
          border: '#0B0D12',
          fallback: '#EEF1F5'
        },
        tertiary: {
          default: '#DDE3EA',
          border: '#0B0D12',
          fallback: '#DDE3EA'
        }
      },
      ui: {
        primary: {
          default: '#111318',
          background: 'rgba(17, 19, 24, 0.08)',
          inactive: 'rgba(17, 19, 24, 0.5)',
          fallback: '#111318'
        },
        secondary: {
          default: '#334155',
          background: 'rgba(51, 65, 85, 0.1)',
          inactive: 'rgba(51, 65, 85, 0.5)',
          fallback: '#334155'
        },
        tertiary: '#8A96A8'
      },
      text: {
        primary: '#050505',
        secondary: '#485465'
      }
    },
    browser: {
      themeColor: '#006B8F',
      selectionText: '#FFFFFF',
      selectionBackground: '#006B8F',
      caret: '#006B8F',
      scrollbarThumb: '#0B0D12',
      scrollbarThumbHover: '#006B8F',
      scrollbarTrack: '#FFFFFF'
    },
    markdown: {
      extends: 'light',
      linkText: '#006B8F',
      quoteBar: '#0B0D12',
      codeBackground: '#EEF1F5',
      embedLinkHighlightText: '#050505',
      embedLinkHighlightBackground: '#C9F3FF'
    }
  }

  /**
   * @type {import('/@/types/theme').Theme}
   */
  const darkTheme = {
    version: 2,
    basic: {
      accent: {
        primary: {
          default: '#7ED6E2',
          background: '#7ED6E2',
          inactive: 'rgba(126, 214, 226, 0.55)',
          fallback: '#7ED6E2'
        },
        notification: {
          default: '#FFB000',
          background: '#FFB000',
          fallback: '#FFB000'
        },
        online: '#38FF7A',
        error: '#FF4560',
        focus: '#7ED6E2C0'
      },
      background: {
        primary: {
          default: '#05070A',
          border: '#F6F7FA',
          fallback: '#05070A'
        },
        secondary: {
          default: '#0D1117',
          border: '#F6F7FA',
          fallback: '#0D1117'
        },
        tertiary: {
          default: '#171D26',
          border: '#F6F7FA',
          fallback: '#171D26'
        }
      },
      ui: {
        primary: {
          default: '#F6F7FA',
          background: 'rgba(246, 247, 250, 0.12)',
          inactive: 'rgba(246, 247, 250, 0.55)',
          fallback: '#F6F7FA'
        },
        secondary: {
          default: '#B8C3D6',
          background: 'rgba(184, 195, 214, 0.12)',
          inactive: 'rgba(184, 195, 214, 0.55)',
          fallback: '#B8C3D6'
        },
        tertiary: '#6F7D94'
      },
      text: {
        primary: '#FFFFFF',
        secondary: '#C4CDD9'
      }
    },
    specific: {
      stampEdgeEnable: true
    },
    browser: {
      themeColor: '#05070A',
      colorScheme: 'dark',
      selectionText: '#05070A',
      selectionBackground: '#7ED6E2',
      caret: '#7ED6E2',
      scrollbarThumb: '#F6F7FA',
      scrollbarThumbHover: '#7ED6E2',
      scrollbarTrack: '#05070A'
    },
    markdown: {
      extends: 'dark',
      linkText: '#7ED6E2',
      quoteBar: '#F6F7FA',
      codeBackground: '#171D26',
      embedLinkHighlightText: '#05070A',
      embedLinkHighlightBackground: '#7CFF00'
    }
  }

  window.defaultLightTheme = lightTheme
  window.defaultDarkTheme = darkTheme
})()
