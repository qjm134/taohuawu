// 背景元素类
class Background extends Phaser.GameObjects.Container {
    constructor(scene) {
        super(scene, 0, 0);

        this.scene = scene;
        this.scene.add.existing(this);

        // 创建天空
        this.createSky();

        // 创建水面
        this.createRiver();

        // 创建桥梁
        this.createBridge();

        // 创建青石板路
        this.createStoneRoad();

        // 创建两岸
        this.createBanks();

        // 创建装饰元素
        this.createDecorations();

        // 创建乌篷船
        this.createBoat();
    }

    createSky() {
        // 天空渐变背景
        const sky = this.scene.add.graphics();
        sky.fillGradientStyle(0x87CEEB, 0x87CEEB, 0xB0E0E6, 0xB0E0E6, 1);
        sky.fillRect(0, 0, this.scene.scale.width, this.scene.scale.height * 0.4);
        this.add(sky);

        // 添加一些云朵
        for (let i = 0; i < 5; i++) {
            const cloud = this.createCloud(
                Phaser.Math.Between(50, this.scene.scale.width - 50),
                Phaser.Math.Between(20, 100)
            );
            this.add(cloud);
        }
    }

    createCloud(x, y) {
        const cloud = this.scene.add.graphics();
        const size = Phaser.Math.Between(30, 50);

        cloud.fillStyle(0xFFFFFF, 0.9);
        cloud.fillCircle(0, 0, size);
        cloud.fillCircle(-size * 0.6, size * 0.3, size * 0.7);
        cloud.fillCircle(size * 0.6, size * 0.3, size * 0.7);
        cloud.fillCircle(-size * 0.3, -size * 0.3, size * 0.6);
        cloud.fillCircle(size * 0.3, -size * 0.3, size * 0.6);

        cloud.setPosition(x, y);

        // 云朵飘动动画
        this.scene.tweens.add({
            targets: cloud,
            x: x + Phaser.Math.Between(-20, 20),
            duration: Phaser.Math.Between(10000, 20000),
            yoyo: true,
            repeat: -1,
            ease: 'Sine.easeInOut',
        });

        return cloud;
    }

    createRiver() {
        // 河流区域
        const river = this.scene.add.graphics();
        const riverY = this.scene.scale.height * 0.5;

        // 河流主体
        river.fillStyle(0x4A90A4, 1);
        river.fillRect(0, riverY - 50, this.scene.scale.width, 100);

        // 添加水波纹效果
        const waves = this.scene.add.graphics();
        waves.lineStyle(2, 0x5BA6BD, 0.5);

        for (let i = 0; i < 20; i++) {
            const waveX = Phaser.Math.Between(0, this.scene.scale.width);
            const waveY = riverY + Phaser.Math.Between(-40, 40);
            const waveWidth = Phaser.Math.Between(50, 150);

            waves.beginPath();
            waves.moveTo(waveX, waveY);
            for (let j = 0; j <= waveWidth; j += 5) {
                const y = waveY + Math.sin(j * 0.05 + i) * 3;
                waves.lineTo(waveX + j, y);
            }
            waves.strokePath();
        }

        // 水波纹动画
        this.scene.tweens.add({
            targets: waves,
            alpha: 0.3,
            duration: 2000,
            yoyo: true,
            repeat: -1,
            ease: 'Sine.easeInOut',
        });

        this.add([river, waves]);
    }

    createBridge() {
        const bridge = this.scene.add.graphics();
        const bridgeY = this.scene.scale.height * 0.5;

        // 桥面
        bridge.fillStyle(0x8B4513, 1);
        bridge.fillRect(200, bridgeY - 30, 800, 60);

        // 桥栏杆
        bridge.fillStyle(0xA0522D, 1);
        bridge.fillRect(200, bridgeY - 45, 800, 10); // 上栏杆
        bridge.fillRect(200, bridgeY + 35, 800, 10); // 下栏杆

        // 桥柱
        const pillarPositions = [250, 400, 550, 700, 850, 950];
        pillarPositions.forEach(x => {
            bridge.fillStyle(0x654321, 1);
            bridge.fillRect(x - 10, bridgeY - 50, 20, 100);
        });

        // 桥拱装饰
        bridge.fillStyle(0xDEB887, 1);
        bridge.fillCircle(600, bridgeY, 20);

        this.add(bridge);
    }

    createStoneRoad() {
        const road = this.scene.add.graphics();
        const roadWidth = 150;
        const roadX = this.scene.scale.width / 2;

        // 道路
        road.fillStyle(0x808080, 1);
        road.fillRect(roadX - roadWidth / 2, this.scene.scale.height * 0.5 + 50, roadWidth, this.scene.scale.height * 0.4);

        // 青石板纹理
        road.fillStyle(0xA9A9A9, 1);
        for (let i = 0; i < 30; i++) {
            const stoneX = roadX + Phaser.Math.Between(-roadWidth / 2 + 10, roadWidth / 2 - 10);
            const stoneY = this.scene.scale.height * 0.5 + 60 + i * 15;
            const stoneWidth = Phaser.Math.Between(20, 40);
            const stoneHeight = Phaser.Math.Between(8, 12);

            road.fillRect(stoneX - stoneWidth / 2, stoneY, stoneWidth, stoneHeight);
        }

        this.add(road);
    }

    createBanks() {
        const banks = this.scene.add.graphics();
        const riverY = this.scene.scale.height * 0.5;

        // 左岸
        banks.fillStyle(0x228B22, 1);
        banks.fillRect(0, riverY - 40, 200, 40);

        // 右岸
        banks.fillStyle(0x228B22, 1);
        banks.fillRect(1000, riverY - 40, 200, 40);

        // 岸边草丛
        banks.fillStyle(0x32CD32, 1);
        for (let i = 0; i < 10; i++) {
            const grassX = Phaser.Math.Between(10, 190);
            const grassY = riverY - 40;
            banks.fillRect(grassX, grassY, 3, 8);
        }

        for (let i = 0; i < 10; i++) {
            const grassX = Phaser.Math.Between(1010, 1190);
            const grassY = riverY - 40;
            banks.fillRect(grassX, grassY, 3, 8);
        }

        this.add(banks);
    }

    createDecorations() {
        // 添加一些装饰性元素
        const decorations = [];

        // 柳树
        for (let i = 0; i < 2; i++) {
            const willow = this.createWillow(100 + i * 1000, this.scene.scale.height * 0.5 - 100);
            decorations.push(willow);
        }

        // 灯笼
        for (let i = 0; i < 6; i++) {
            const lantern = this.createLantern(250 + i * 140, this.scene.scale.height * 0.5 - 50);
            decorations.push(lantern);
        }

        this.add(decorations);
    }

    createWillow(x, y) {
        const willow = this.scene.add.graphics();

        // 树干
        willow.fillStyle(0x8B4513, 1);
        willow.fillRect(-10, 0, 20, 80);

        // 树冠
        willow.fillStyle(0x228B22, 1);
        willow.fillCircle(0, -30, 60);

        // 柳枝
        for (let i = 0; i < 12; i++) {
            const angle = Math.PI * i / 6;
            const branchX = Math.cos(angle) * 30;
            const branchY = Math.sin(angle) * 30 - 30;

            willow.lineStyle(3, 0x32CD32, 1);
            willow.beginPath();
            willow.moveTo(branchX, branchY);
            willow.lineTo(branchX + Math.cos(angle + Math.PI / 2) * 40, branchY + 40);
            willow.strokePath();
        }

        willow.setPosition(x, y);
        return willow;
    }

    createLantern(x, y) {
        const lantern = this.scene.add.graphics();

        // 灯笼主体
        lantern.fillStyle(0xFF4500, 1);
        lantern.fillEllipse(0, 0, 20, 30);

        // 灯笼边缘
        lantern.lineStyle(2, 0xFFD700, 1);
        lantern.strokeEllipse(0, 0, 20, 30);

        // 灯笼提手
        lantern.lineStyle(2, 0xFFD700, 1);
        lantern.beginPath();
        lantern.moveTo(-5, -15);
        lantern.lineTo(-5, -25);
        lantern.lineTo(5, -25);
        lantern.lineTo(5, -15);
        lantern.strokePath();

        lantern.setPosition(x, y);

        // 灯笼摇摆动画
        this.scene.tweens.add({
            targets: lantern,
            rotation: Phaser.Math.DegToRad(5),
            duration: 2000,
            yoyo: true,
            repeat: -1,
            ease: 'Sine.easeInOut',
        });

        return lantern;
    }

    createBoat() {
        const boat = this.scene.add.graphics();

        // 船身
        boat.fillStyle(0x654321, 1);
        boat.fillEllipse(0, 0, 100, 30);

        // 船篷
        boat.fillStyle(0x228B22, 1);
        boat.fillEllipse(0, -20, 80, 20);

        // 船桨
        boat.lineStyle(3, 0x8B4513, 1);
        boat.beginPath();
        boat.moveTo(30, 5);
        boat.lineTo(60, 20);
        boat.strokePath();

        boat.setPosition(this.scene.scale.width / 2, this.scene.scale.height * 0.5);

        // 船只漂浮动画
        this.scene.tweens.add({
            targets: boat,
            y: boat.y + 10,
            duration: 2000,
            yoyo: true,
            repeat: -1,
            ease: 'Sine.easeInOut',
        });

        this.add(boat);
    }
}