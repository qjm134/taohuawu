// NPC 导游类
class NPCGuide extends Phaser.GameObjects.Container {
    constructor(scene, x, y, options = {}) {
        super(scene, x, y);

        this.scene = scene;
        this.scene.add.existing(this);

        // 配置
        this.name = options.name || '小荷';
        this.scale = options.scale || 1;
        this.color = options.color || 0xFF69B4;

        // 创建身体
        this.body = scene.add.ellipse(0, 0, 40 * this.scale, 80 * this.scale, this.color);
        this.body.setStrokeStyle(2, 0x000000);

        // 创建头部
        this.head = scene.add.circle(0, -45 * this.scale, 25 * this.scale, 0xFFE0BD);
        this.head.setStrokeStyle(2, 0x000000);

        // 创建眼睛
        this.leftEye = scene.add.circle(-8 * this.scale, -48 * this.scale, 4 * this.scale, 0x000000);
        this.rightEye = scene.add.circle(8 * this.scale, -48 * this.scale, 4 * this.scale, 0x000000);

        // 创建嘴巴（微笑）
        this.mouth = scene.add.arc(0, -38 * this.scale, 8 * this.scale, 0, Math.PI, false, 0xFF6B6B);
        this.mouth.setStrokeStyle(2, 0x000000);

        // 创建头发
        this.hairLeft = scene.add.ellipse(-15 * this.scale, -55 * this.scale, 15 * this.scale, 20 * this.scale, 0x8B4513);
        this.hairRight = scene.add.ellipse(15 * this.scale, -55 * this.scale, 15 * this.scale, 20 * this.scale, 0x8B4513);
        this.hairTop = scene.add.ellipse(0, -60 * this.scale, 25 * this.scale, 15 * this.scale, 0x8B4513);

        // 创建发簪（荷花）
        this.hairpin = scene.add.polygon(
            0, -65 * this.scale,
            [0, -10, 8, 5, -8, 5],
            0xFF69B4
        );
        this.hairpin.setStrokeStyle(2, 0x000000);

        // 添加所有元素到容器
        this.add([
            this.body, this.head, this.leftEye, this.rightEye, this.mouth,
            this.hairLeft, this.hairRight, this.hairTop, this.hairpin,
        ]);

        // 设置可交互
        this.setSize(60 * this.scale, 100 * this.scale);
        this.setInteractive({ cursor: 'pointer' });

        // 添加阴影
        this.shadow = scene.add.ellipse(0, 50 * this.scale, 50 * this.scale, 10 * this.scale, 0x000000, 0.2);

        // 状态
        this.isBreathing = true;
        this.breathTween = null;
        this.currentEmotion = 'neutral';

        // 启动呼吸动画
        this.startBreathing();
    }

    startBreathing() {
        if (this.breathTween) {
            this.breathTween.remove();
        }

        this.breathTween = this.scene.tweens.add({
            targets: this,
            scale: this.scale * 1.02,
            yoyo: true,
            repeat: -1,
            duration: 2000,
            ease: 'Sine.easeInOut',
        });
    }

    stopBreathing() {
        if (this.breathTween) {
            this.breathTween.remove();
            this.breathTween = null;
        }
        this.setScale(this.scale);
    }

    setEmotion(emotion) {
        this.currentEmotion = emotion;

        // 根据情绪调整表情
        switch (emotion) {
            case 'happy':
                this.updateMouth(Math.PI * 0.7, 0x4CAF50);
                break;
            case 'sad':
                this.updateMouth(Math.PI, 0xFF6B6B);
                this.body.tint = 0x999999;
                break;
            case 'angry':
                this.updateMouth(Math.PI * 1.3, 0xFF4444);
                this.body.tint = 0xFF6666;
                break;
            case 'confused':
                this.updateMouth(Math.PI * 0.3, 0xFFA500);
                break;
            case 'excited':
                this.updateMouth(Math.PI, 0xFFD700);
                this.body.tint = 0xFFFFAA;
                break;
            default:
                this.updateMouth(Math.PI, 0xFF6B6B);
                this.body.clearTint();
        }
    }

    updateMouth(startAngle, color) {
        this.mouth.clear();
        this.mouth.beginPath();
        this.mouth.arc(0, -38 * this.scale, 8 * this.scale, startAngle, startAngle + Math.PI, false);
        this.mouth.strokePath();
        this.moth.fillStyle(color, 1);
        this.mouth.fillPath();
    }

    wave() {
        this.stopBreathing();

        // 挥手动画
        this.scene.tweens.add({
            targets: this.hairpin,
            rotation: -0.5,
            duration: 200,
            yoyo: true,
            repeat: 3,
            onComplete: () => {
                this.hairpin.rotation = 0;
                this.startBreathing();
            }
        });
    }

    jump() {
        this.stopBreathing();

        this.scene.tweens.add({
            targets: this,
            y: this.y - 30,
            duration: 200,
            ease: 'Quad.easeOut',
            yoyo: true,
            repeat: 0,
            onComplete: () => {
                this.startBreathing();
            }
        });
    }

    destroy() {
        this.stopBreathing();
        super.destroy();
    }
}