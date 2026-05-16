export namespace inventory {
	
	export class RentalSummary {
	    totalDays: number;
	    totalIncome: number;
	    rentCount: number;
	
	    static createFrom(source: any = {}) {
	        return new RentalSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalDays = source["totalDays"];
	        this.totalIncome = source["totalIncome"];
	        this.rentCount = source["rentCount"];
	    }
	}
	export class ItemDetail {
	    item: model.InventoryItem;
	    rentalHistory: model.RentalRecord[];
	    rentalSummary: RentalSummary;
	
	    static createFrom(source: any = {}) {
	        return new ItemDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item = this.convertValues(source["item"], model.InventoryItem);
	        this.rentalHistory = this.convertValues(source["rentalHistory"], model.RentalRecord);
	        this.rentalSummary = this.convertValues(source["rentalSummary"], RentalSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class DashboardSummary {
	    totalNetWorth: number;
	    inventoryCount: number;
	    completedTrades: number;
	    totalRentalIncome: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalNetWorth = source["totalNetWorth"];
	        this.inventoryCount = source["inventoryCount"];
	        this.completedTrades = source["completedTrades"];
	        this.totalRentalIncome = source["totalRentalIncome"];
	    }
	}

}

export namespace model {
	
	export class Account {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    platform: string;
	    availableBalance: number;
	    purchaseBalance: number;
	    remark: string;
	    status: string;
	    lastSyncAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.platform = source["platform"];
	        this.availableBalance = source["availableBalance"];
	        this.purchaseBalance = source["purchaseBalance"];
	        this.remark = source["remark"];
	        this.status = source["status"];
	        this.lastSyncAt = source["lastSyncAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TradeRecord {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    assetId?: string;
	    classId?: string;
	    instanceId?: string;
	    goodsId?: number;
	    marketHashName?: string;
	    itemName?: string;
	    iconUrl?: string;
	    paintWear?: number;
	    paintSeed?: number;
	    paintIndex?: number;
	    categoryGroup?: string;
	    exterior?: string;
	    rarity?: string;
	    weaponType?: string;
	    weaponName?: string;
	    quality?: string;
	    series?: string;
	    itemset?: string;
	    weaponCase?: string;
	    custom?: string;
	    stickers?: Sticker[];
	    keychains?: Keychain[];
	    tradableUnfrozenTime?: number;
	    accountId: number;
	    tradeType: string;
	    quantity: number;
	    unitPrice: number;
	    totalPrice: number;
	    fee: number;
	    tradeAt: number;
	    source: string;
	    state: string;
	    stateText: string;
	    transactTime?: number;
	    tradeOfferId: string;
	    externalId: string;
	    matchedBuyTradeId?: number;
	    consumedQuantity: number;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new TradeRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.assetId = source["assetId"];
	        this.classId = source["classId"];
	        this.instanceId = source["instanceId"];
	        this.goodsId = source["goodsId"];
	        this.marketHashName = source["marketHashName"];
	        this.itemName = source["itemName"];
	        this.iconUrl = source["iconUrl"];
	        this.paintWear = source["paintWear"];
	        this.paintSeed = source["paintSeed"];
	        this.paintIndex = source["paintIndex"];
	        this.categoryGroup = source["categoryGroup"];
	        this.exterior = source["exterior"];
	        this.rarity = source["rarity"];
	        this.weaponType = source["weaponType"];
	        this.weaponName = source["weaponName"];
	        this.quality = source["quality"];
	        this.series = source["series"];
	        this.itemset = source["itemset"];
	        this.weaponCase = source["weaponCase"];
	        this.custom = source["custom"];
	        this.stickers = this.convertValues(source["stickers"], Sticker);
	        this.keychains = this.convertValues(source["keychains"], Keychain);
	        this.tradableUnfrozenTime = source["tradableUnfrozenTime"];
	        this.accountId = source["accountId"];
	        this.tradeType = source["tradeType"];
	        this.quantity = source["quantity"];
	        this.unitPrice = source["unitPrice"];
	        this.totalPrice = source["totalPrice"];
	        this.fee = source["fee"];
	        this.tradeAt = source["tradeAt"];
	        this.source = source["source"];
	        this.state = source["state"];
	        this.stateText = source["stateText"];
	        this.transactTime = source["transactTime"];
	        this.tradeOfferId = source["tradeOfferId"];
	        this.externalId = source["externalId"];
	        this.matchedBuyTradeId = source["matchedBuyTradeId"];
	        this.consumedQuantity = source["consumedQuantity"];
	        this.remark = source["remark"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Keychain {
	    name?: string;
	    imageUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new Keychain(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.imageUrl = source["imageUrl"];
	    }
	}
	export class Sticker {
	    stickerId?: number;
	    slot?: number;
	    wear?: number;
	    name?: string;
	    imageUrl?: string;
	    referencePrice?: number;
	    offsetX?: number;
	    offsetY?: number;
	
	    static createFrom(source: any = {}) {
	        return new Sticker(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stickerId = source["stickerId"];
	        this.slot = source["slot"];
	        this.wear = source["wear"];
	        this.name = source["name"];
	        this.imageUrl = source["imageUrl"];
	        this.referencePrice = source["referencePrice"];
	        this.offsetX = source["offsetX"];
	        this.offsetY = source["offsetY"];
	    }
	}
	export class InventoryItem {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    assetId?: string;
	    classId?: string;
	    instanceId?: string;
	    goodsId?: number;
	    marketHashName?: string;
	    itemName?: string;
	    iconUrl?: string;
	    paintWear?: number;
	    paintSeed?: number;
	    paintIndex?: number;
	    categoryGroup?: string;
	    exterior?: string;
	    rarity?: string;
	    weaponType?: string;
	    weaponName?: string;
	    quality?: string;
	    series?: string;
	    itemset?: string;
	    weaponCase?: string;
	    custom?: string;
	    stickers?: Sticker[];
	    keychains?: Keychain[];
	    tradableUnfrozenTime?: number;
	    accountId: number;
	    buyTradeId: number;
	    quantity: number;
	    status: string;
	    listedPrice?: number;
	    listedAt?: number;
	    buyTrade?: TradeRecord;
	
	    static createFrom(source: any = {}) {
	        return new InventoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.assetId = source["assetId"];
	        this.classId = source["classId"];
	        this.instanceId = source["instanceId"];
	        this.goodsId = source["goodsId"];
	        this.marketHashName = source["marketHashName"];
	        this.itemName = source["itemName"];
	        this.iconUrl = source["iconUrl"];
	        this.paintWear = source["paintWear"];
	        this.paintSeed = source["paintSeed"];
	        this.paintIndex = source["paintIndex"];
	        this.categoryGroup = source["categoryGroup"];
	        this.exterior = source["exterior"];
	        this.rarity = source["rarity"];
	        this.weaponType = source["weaponType"];
	        this.weaponName = source["weaponName"];
	        this.quality = source["quality"];
	        this.series = source["series"];
	        this.itemset = source["itemset"];
	        this.weaponCase = source["weaponCase"];
	        this.custom = source["custom"];
	        this.stickers = this.convertValues(source["stickers"], Sticker);
	        this.keychains = this.convertValues(source["keychains"], Keychain);
	        this.tradableUnfrozenTime = source["tradableUnfrozenTime"];
	        this.accountId = source["accountId"];
	        this.buyTradeId = source["buyTradeId"];
	        this.quantity = source["quantity"];
	        this.status = source["status"];
	        this.listedPrice = source["listedPrice"];
	        this.listedAt = source["listedAt"];
	        this.buyTrade = this.convertValues(source["buyTrade"], TradeRecord);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RentalRecord {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    accountId: number;
	    assetId: string;
	    itemName: string;
	    income: number;
	    durationDays: number;
	    startAt: number;
	    endAt: number;
	    externalId: string;
	
	    static createFrom(source: any = {}) {
	        return new RentalRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.accountId = source["accountId"];
	        this.assetId = source["assetId"];
	        this.itemName = source["itemName"];
	        this.income = source["income"];
	        this.durationDays = source["durationDays"];
	        this.startAt = source["startAt"];
	        this.endAt = source["endAt"];
	        this.externalId = source["externalId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace pnl {
	
	export class MonthlyPLView {
	    month: string;
	    netPl: number;
	
	    static createFrom(source: any = {}) {
	        return new MonthlyPLView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.month = source["month"];
	        this.netPl = source["netPl"];
	    }
	}
	export class PnlSummaryView {
	    totalTrades: number;
	    totalGrossPl: number;
	    totalFee: number;
	    totalNetPl: number;
	
	    static createFrom(source: any = {}) {
	        return new PnlSummaryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalTrades = source["totalTrades"];
	        this.totalGrossPl = source["totalGrossPl"];
	        this.totalFee = source["totalFee"];
	        this.totalNetPl = source["totalNetPl"];
	    }
	}

}

export namespace sync {
	
	export class SyncResult {
	    NewTrades: number;
	    NewPnl: number;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new SyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.NewTrades = source["NewTrades"];
	        this.NewPnl = source["NewPnl"];
	        this.warnings = source["warnings"];
	    }
	}

}

export namespace trade {
	
	export class CompletedTradeView {
	    itemName: string;
	    exterior: string;
	    paintWear: number;
	    quantity: number;
	    grossPl: number;
	    totalFee: number;
	    netPl: number;
	    sellTrade: model.TradeRecord;
	    buyTrade: model.TradeRecord;
	
	    static createFrom(source: any = {}) {
	        return new CompletedTradeView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemName = source["itemName"];
	        this.exterior = source["exterior"];
	        this.paintWear = source["paintWear"];
	        this.quantity = source["quantity"];
	        this.grossPl = source["grossPl"];
	        this.totalFee = source["totalFee"];
	        this.netPl = source["netPl"];
	        this.sellTrade = this.convertValues(source["sellTrade"], model.TradeRecord);
	        this.buyTrade = this.convertValues(source["buyTrade"], model.TradeRecord);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CompletedTradesSummary {
	    totalTrades: number;
	    totalGrossPl: number;
	    totalFee: number;
	    totalNetPl: number;
	
	    static createFrom(source: any = {}) {
	        return new CompletedTradesSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalTrades = source["totalTrades"];
	        this.totalGrossPl = source["totalGrossPl"];
	        this.totalFee = source["totalFee"];
	        this.totalNetPl = source["totalNetPl"];
	    }
	}

}

