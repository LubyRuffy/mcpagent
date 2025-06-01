/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '@element-plus/icons-vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
  export const Globe: DefineComponent<{}, {}, any>
  export const Sunny: DefineComponent<{}, {}, any>
  export const Moon: DefineComponent<{}, {}, any>
  export const Monitor: DefineComponent<{}, {}, any>
  export const Connection: DefineComponent<{}, {}, any>
  export const Expand: DefineComponent<{}, {}, any>
  export const Fold: DefineComponent<{}, {}, any>
  export const MoreFilled: DefineComponent<{}, {}, any>
  export const DocumentAdd: DefineComponent<{}, {}, any>
  export const Download: DefineComponent<{}, {}, any>
  export const Upload: DefineComponent<{}, {}, any>
  export const RefreshLeft: DefineComponent<{}, {}, any>
  export const User: DefineComponent<{}, {}, any>
  export const Robot: DefineComponent<{}, {}, any>
  export const DocumentCopy: DefineComponent<{}, {}, any>
  export const BrainFilled: DefineComponent<{}, {}, any>
  export const WarningFilled: DefineComponent<{}, {}, any>
  export const InfoFilled: DefineComponent<{}, {}, any>
  export const ChatDotRound: DefineComponent<{}, {}, any>
  export const Tools: DefineComponent<{}, {}, any>
  export const ArrowDown: DefineComponent<{}, {}, any>
  export const ArrowUp: DefineComponent<{}, {}, any>
  export const Paperclip: DefineComponent<{}, {}, any>
  export const Clock: DefineComponent<{}, {}, any>
  export const Delete: DefineComponent<{}, {}, any>
  export const Promotion: DefineComponent<{}, {}, any>
  export const View: DefineComponent<{}, {}, any>
}
