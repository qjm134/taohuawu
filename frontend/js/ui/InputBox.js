// 输入框组件
class InputBox extends Phaser.GameObjects.Container {
    constructor(scene, x, y, options = {}) {
        super(scene, x, y);

        this.scene = scene;
        this.scene.add.existing(this);

        // 配置
        this.width = options.width || 600;
        this.height = options.height || 50;
        this.placeholder = options.placeholder || '输入问题向导游小荷提问...';
        this.onSend = options.onSend || null;

        // 创建输入框背景
        this.background = scene.add.graphics();
        this.background.fillStyle(0xFFFFFF, 0.95);
        this.background.lineStyle(2, 0xCCCCCC);
        this.background.fillRoundedRect(-this.width / 2, -this.height / 2, this.width, this.height, 10);
        this.background.strokeRoundedRect(-this.width / 2, -this.height / 2, this.width, this.height, 10);

        // 创建 DOM 输入元素
        const inputElement = document.createElement('input');
        inputElement.type = 'text';
        inputElement.placeholder = this.placeholder;
        inputElement.style.cssText = `
            position: absolute;
            width: ${this.width - 20}px;
            height: ${this.height - 10}px;
            font-family: 'Microsoft YaHei', sans-serif;
            font-size: 14px;
            padding: 5px 10px;
            border: none;
            outline: none;
            background: transparent;
            color: #333333;
        `;

        // 创建发送按钮
        this.sendButton = scene.add.text(
            this.width / 2 - 60,
            0,
            '发送',
            {
                fontFamily: 'Microsoft YaHei',
                fontSize: '16px',
                color: '#FF69B4',
                fontStyle: 'bold',
            }
        );
        this.sendButton.setOrigin(0.5);
        this.sendButton.setInteractive({ useHandCursor: true });

        // 添加所有元素到容器
        this.add([this.background, this.sendButton]);

        // DOM 元素
        this.inputElement = inputElement;

        // 事件监听
        this.setupEvents();

        // 初始隐藏
        this.setVisible(false);
    }

    setupEvents() {
        // 发送按钮点击
        this.sendButton.on('pointerdown', () => {
            this.send();
        });

        // 输入框回车
        this.inputElement.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                this.send();
            }
        });
    }

    send() {
        const text = this.inputElement.value.trim();
        if (text && this.onSend) {
            this.onSend(text);
            this.inputElement.value = '';
        }
    }

    show() {
        this.setVisible(true);

        // 添加 DOM 元素
        const bounds = this.getBounds();
        this.scene.add.existing(this);

        this.inputElement.style.left = (bounds.x + 10) + 'px';
        this.inputElement.style.top = (bounds.y + 5) + 'px';
        document.body.appendChild(this.inputElement);

        // 聚焦
        setTimeout(() => {
            this.inputElement.focus();
        }, 100);
    }

    hide() {
        this.setVisible(false);

        // 移除 DOM 元素
        if (this.inputElement && this.inputElement.parentNode) {
            this.inputElement.parentNode.removeChild(this.inputElement);
        }
    }

    focus() {
        if (this.inputElement) {
            this.inputElement.focus();
        }
    }

    blur() {
        if (this.inputElement) {
            this.inputElement.blur();
        }
    }

    getValue() {
        return this.inputElement.value;
    }

    setValue(value) {
        this.inputElement.value = value;
    }

    clear() {
        this.inputElement.value = '';
    }

    destroy() {
        if (this.inputElement && this.inputElement.parentNode) {
            this.inputElement.parentNode.removeChild(this.inputElement);
        }
        super.destroy();
    }
}