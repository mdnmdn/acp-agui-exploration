import * as React from 'react';
import * as ReactDOM from 'react-dom/client';
import { CopilotSidebar, CopilotPopup, CopilotChat } from "@copilotkit/react-ui";
import { CopilotKit } from "@copilotkit/react-core";
import "@copilotkit/react-ui/styles.css";

export interface CopilotKitEventMap {
  'copilotkit:open': CustomEvent<void>;
  'copilotkit:close': CustomEvent<void>;
  'copilotkit:message-sent': CustomEvent<{ message: string }>;
  'copilotkit:feedback': CustomEvent<{ messageId: string; type: string }>;
  'copilotkit:copy': CustomEvent<{ messageId: string }>;
  'copilotkit:regenerate': CustomEvent<{ messageId: string }>;
  'copilotkit:rewrite': CustomEvent<{ messageId: string }>;
  'copilotkit:generation-changed': CustomEvent<{ isGenerating: boolean }>;
}

export interface CopilotKitEventCallbacks {
  onOpen?: () => void;
  onClose?: () => void;
  onMessageSent?: (message: string) => void;
  onFeedback?: (messageId: string, type: string) => void;
  onCopy?: (messageId: string) => void;
  onRegenerate?: (messageId: string) => void;
  onRewrite?: (messageId: string) => void;
  onGenerationChanged?: (isGenerating: boolean) => void;
}

class CopilotKitWebComponent extends HTMLElement {
  private root: ReactDOM.Root | null = null;
  private container: HTMLDivElement | null = null;

  private _runtimeUrl = '';
  private _publicApiKey = '';
  private _mode: 'chat' | 'sidebar' | 'popup' = 'sidebar';
  private _labels = {
    title: 'Your Assistant',
    initial: 'Hi! How can I assist you today?',
    placeholder: 'Type a message...',
  };
  private _defaultOpen = false;
  private _callbacks: CopilotKitEventCallbacks = {};

  static observedAttributes = [
    'url',
    'apikey',
    'mode',
    'default-open',
    'labels-title',
    'labels-initial',
    'labels-placeholder',
  ];

  constructor() {
    super();
  }

  get callbacks(): CopilotKitEventCallbacks {
    return this._callbacks;
  }

  set callbacks(value: CopilotKitEventCallbacks) {
    this._callbacks = value;
    this.render();
  }

  connectedCallback() {
    if (this.container) return;

    this.ensureStyles();

    this.container = document.createElement('div');
    this.container.className = 'acp-copilotkit-container';
    this.appendChild(this.container);

    this.root = ReactDOM.createRoot(this.container);

    this.updateStateFromAttributes();
    this.render();
  }

  disconnectedCallback() {
    if (this.root) {
      this.root.unmount();
      this.root = null;
    }
    if (this.container) {
      this.container.remove();
      this.container = null;
    }
  }

  attributeChangedCallback() {
    this.updateStateFromAttributes();
    this.render();
  }

  private ensureStyles() {
    if (document.getElementById('acp-copilotkit-styles')) return;
    const style = document.createElement('style');
    style.id = 'acp-copilotkit-styles';
    style.textContent = `
      acp-copilotkit {
        display: block;
        width: 100%;
      }
      .acp-copilotkit-container {
        width: 100%;
        height: 100%;
        min-height: 200px;
      }
    `;
    document.head.appendChild(style);
  }

  private updateStateFromAttributes() {
    const url = this.getAttribute('url');
    if (url !== null) this._runtimeUrl = url;

    const apikey = this.getAttribute('apikey');
    if (apikey !== null) this._publicApiKey = apikey;

    const mode = this.getAttribute('mode');
    if (mode === 'chat' || mode === 'sidebar' || mode === 'popup') {
      this._mode = mode;
    }

    const defaultOpen = this.getAttribute('default-open');
    this._defaultOpen = defaultOpen !== null && defaultOpen !== 'false';

    const title = this.getAttribute('labels-title');
    if (title !== null) this._labels.title = title;

    const initial = this.getAttribute('labels-initial');
    if (initial !== null) this._labels.initial = initial;

    const placeholder = this.getAttribute('labels-placeholder');
    if (placeholder !== null) this._labels.placeholder = placeholder;
  }

  private dispatch(eventName: string, detail: unknown) {
    this.dispatchEvent(new CustomEvent(`copilotkit:${eventName}`, { detail }));
  }

  private buildObservabilityHooks() {
    const cb = this._callbacks;
    return {
      onMessageSent: (message: string) => {
        cb.onMessageSent?.(message);
        this.dispatch('message-sent', { message });
      },
      onChatExpanded: () => {
        cb.onOpen?.();
        this.dispatch('open', undefined);
      },
      onChatMinimized: () => {
        cb.onClose?.();
        this.dispatch('close', undefined);
      },
      onFeedbackGiven: (messageId: string, type: string) => {
        cb.onFeedback?.(messageId, type);
        this.dispatch('feedback', { messageId, type });
      },
      onMessageCopied: (messageId: string) => {
        cb.onCopy?.(messageId);
        this.dispatch('copy', { messageId });
      },
      onRegenerate: (messageId: string) => {
        cb.onRegenerate?.(messageId);
        this.dispatch('regenerate', { messageId });
      },
      onRewrite: (messageId: string) => {
        cb.onRewrite?.(messageId);
        this.dispatch('rewrite', { messageId });
      },
      onChatGenerationChanged: (isGenerating: boolean) => {
        cb.onGenerationChanged?.(isGenerating);
        this.dispatch('generation-changed', { isGenerating });
      },
    };
  }

  private render() {
    if (!this.root || !this.container) return;

    const obsHooks = this._publicApiKey ? this.buildObservabilityHooks() : undefined;

    this.root.render(
      <CopilotKit
        runtimeUrl={this._runtimeUrl}
        publicApiKey={this._publicApiKey || undefined}
      >
        {this.renderMode(obsHooks)}
      </CopilotKit>
    );
  }

  private renderMode(obsHooks?: Record<string, (...args: never[]) => void>): React.ReactNode {
    const labels = this._labels;
    const defaultOpen = this._defaultOpen;

    const chatProps = obsHooks ? { observabilityHooks: obsHooks } : {};

    switch (this._mode) {
      case 'sidebar':
        return (
          <CopilotSidebar labels={labels} defaultOpen={defaultOpen} {...chatProps}>
            <></>
          </CopilotSidebar>
        );
      case 'popup':
        return (
          <CopilotPopup labels={labels} defaultOpen={defaultOpen} {...chatProps}>
            <></>
          </CopilotPopup>
        );
      case 'chat':
        return <CopilotChat {...chatProps} />;
      default:
        return (
          <CopilotSidebar labels={labels} defaultOpen={defaultOpen} {...chatProps}>
            <></>
          </CopilotSidebar>
        );
    }
  }
}

customElements.define('acp-copilotkit', CopilotKitWebComponent);

export default CopilotKitWebComponent;
export { CopilotKitWebComponent };
