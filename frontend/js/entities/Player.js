// 玩家类
class Player extends Phaser.GameObjects.Container {
    constructor(scene, x, y, options = {}) {
        super(scene, x, y);

        this.scene = scene;
        this.scene.add.existing(this);

        // 配置
        this.speed = options.speed || 200;
        this.scale = options.scale || 1;

        // 创建身体
        this.body = scene.add.ellipse(0, 0, 35 * this.scale, 70 * this.scale, 0x4169E1);
        this.body.setStrokeStyle(2, 0x000000);

        // 创建头部
        this.head = scene.add.circle(0, -40 * this.scale, 22 * this.scale, 0xFFE0BD);
        this.head.setStrokeStyle(2, 0x000000);

        // 创建眼睛
        this.leftEye = scene.add.circle(-7 * this.scale, -43 * this.scale, 3 * this.scale, 0x000000);
        this.rightEye = scene.add.circle(7 * this.scale, -43 * this.scale, 3 * this.scale, 0x000000);

        // 创建嘴巴
        this.mouth = scene.add.arc(0, -35 * this.scale, 6 * this.scale, 0, Math.PI, false, 0xFF6B6B);
        this.mouth.setStrokeStyle(2, 0x000000);

        // 添加所有元素到容器
        this.add([this.body, this.head, this.leftEye, this.rightEye, this.mouth]);

        // 添加阴影
        this.shadow = scene.add.ellipse(0, 45 * this.scale, 45 * this.scale, 8 * this.scale, 0x000000, 0.2);

        // 移动状态
        this.isMoving = false;
        this.targetX = x;
        this.targetY = y;

        // 键盘控制
        this.cursors = scene.input.keyboard.createCursorKeys();
        this.wasd = scene.input.keyboard.addKeys({
            w: Phaser.Input.Keyboard.KeyCodes.W,
            a: Phaser.Input.Keyboard.KeyCodes.A,
            s: Phaser.Input.Keyboard.KeyCodes.S,
            d: Phaser.Input.Keyboard.KeyCodes.D,
        });
    }

    update(delta) {
        // 处理键盘输入
        let vx = 0;
        let vy = 0;

        if (this.cursors.left.isDown || this.wasd.a.isDown) {
            vx = -this.speed;
        } else if (this.cursors.right.isDown || this.wasd.d.isDown) {
            vx = this.speed;
        }

        if (this.cursors.up.isDown || this.wasd.w.isDown) {
            vy = -this.speed;
        } else if (this.cursors.down.isDown || this.wasd.s.isDown) {
            vy = this.speed;
        }

        // 应用速度
        if (vx !== 0 || vy !== 0) {
            this.isMoving = true;

            // 更新位置
            this.x += vx * (delta / 1000);
            this.y += vy * (delta / 1000);

            // 限制在场景边界内
            this.clampPosition();
        } else {
            this.isMoving = false;
        }

        // 更新阴影位置
        this.shadow.x = this.x;
        this.shadow.y = this.y + 45 * this.scale;
    }

    clampPosition() {
        const margin = 50;
        this.x = Phaser.Math.Clamp(
            this.x,
            margin,
            this.scene.scale.width - margin
        );
        this.y = Phaser.Math.Clamp(
            this.y,
            margin,
            this.scene.scale.height - margin
        );
    }

    moveTo(x, y) {
        this.targetX = x;
        this.targetY = y;

        // 计算距离和角度
        const dx = x - this.x;
        const dy = y - this.y;
        const distance = Math.sqrt(dx * dx + dy * dy);

        if (distance > 0) {
            const duration = (distance / this.speed) * 1000;

            this.scene.tweens.add({
                targets: this,
                x: x,
                y: y,
                duration: duration,
                ease: 'Linear',
                onUpdate: () => {
                    this.isMoving = true;
                    this.shadow.x = this.x;
                    this.shadow.y = this.y + 45 * this.scale;
                },
                onComplete: () => {
                    this.isMoving = false;
                },
            });
        }
    }

    destroy() {
        if (this.shadow) {
            this.shadow.destroy();
        }
        super.destroy();
    }
}