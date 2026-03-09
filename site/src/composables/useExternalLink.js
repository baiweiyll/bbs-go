/**
 * 站外链接跳转前确认
 * 当系统设置开启「在跳转前需手动确认是否前往该站外链接」时，对站外链接进行拦截并弹窗确认
 */

export function isExternalUrl(url) {
  if (!url || typeof url !== "string") return false;
  const u = url.trim();
  return u.startsWith("http://") || u.startsWith("https://");
}

/**
 * 返回链接点击处理函数：若为站外链接且已开启确认，则先弹窗再跳转。
 * 必须在 setup 中调用，例如：const handleLinkClick = useHandleExternalLinkClick();
 * @returns {(e: Event, url: string, openInNewTab?: boolean) => void}
 */
export function useHandleExternalLinkClick() {
  const configStore = useConfigStore();
  const { t } = useI18n();

  return function handleExternalLinkClick(e, url, openInNewTab = true) {
    if (!url || !isExternalUrl(url)) return;

    if (configStore.config.urlRedirect) {
      e.preventDefault();
      const message =
        t("pages.redirect.externalConfirm") +
        "\n\n" +
        t("pages.redirect.externalConfirmUrl") +
        url;
      if (confirm(message)) {
        if (openInNewTab) {
          window.open(url, "_blank", "noopener,noreferrer");
        } else {
          window.location.href = url;
        }
      }
    }
  };
}

/**
 * 用于富文本内容区域（v-html）的点击委托：拦截其中的站外链接并弹窗确认
 * 必须在 setup 中调用，在文章/帖子内容容器上使用 @click="handleContentLinkClick"
 */
export function useContentLinkClickHandler() {
  const configStore = useConfigStore();
  const { t } = useI18n();

  return function handleContentLinkClick(e) {
    const anchor = e.target.closest("a");
    if (!anchor || !anchor.href) return;

    const url = anchor.getAttribute("href");
    if (!isExternalUrl(url)) return;

    if (configStore.config.urlRedirect) {
      e.preventDefault();
      e.stopPropagation();
      const message =
        t("pages.redirect.externalConfirm") +
        "\n\n" +
        t("pages.redirect.externalConfirmUrl") +
        url;
      if (confirm(message)) {
        window.open(url, "_blank", "noopener,noreferrer");
      }
    }
  };
}
