#!/usr/bin/env bash
# Publish Indonesian Snacks catalog to the BPP (which forwards to CDS).
# Usage: bash publish-indonesian-snacks.sh [BPP_URL]
# Default BPP_URL = http://localhost:8080

BPP_URL="${1:-https://bppapp.remiges.tech}"

curl -s -X POST "${BPP_URL}/api/v1/catalog/publish" \
  -H "Content-Type: application/json" \
  -d '{
  "catalogs": [
    {
      "id": "cat-indonesia-snacks-001",
      "descriptor": {
        "name": "Toko Camilan Nusantara (Indonesian Snacks Store)",
        "shortDesc": "Camilan autentik Indonesia: biskuit, keripik, dan kue (Authentic Indonesian snacks: biscuits, chips, and cakes)",
        "longDesc": "Koleksi lengkap camilan khas Indonesia dari berbagai daerah. Mulai dari biskuit kaleng ikonik, keripik dengan bumbu nusantara, hingga kue tradisional yang lezat. (Complete collection of authentic Indonesian snacks from various regions. From iconic tin biscuits, chips with archipelago seasoning, to delicious traditional cakes.)"
      },
      "provider": {
        "id": "provider-toko-nusantara-001",
        "descriptor": {
          "name": "Toko Camilan Nusantara",
          "shortDesc": "Distributor camilan Indonesia terpercaya (Trusted Indonesian snack distributor)"
        },
        "availableAt": [
          {
            "geo": { "type": "Point", "coordinates": [106.8456, -6.2088] },
            "address": {
              "streetAddress": "Jl. Sudirman No. 1",
              "addressLocality": "Jakarta Pusat",
              "addressRegion": "DKI Jakarta",
              "postalCode": "10220",
              "addressCountry": "ID"
            }
          }
        ]
      },
      "validity": {
        "startDate": "2026-04-11T00:00:00Z",
        "endDate": "2027-04-11T00:00:00Z"
      },
      "resources": [
        {
          "id": "res-biscuit-khong-guan-001",
          "descriptor": {
            "name": "Biskuit Kaleng Khong Guan (Khong Guan Assorted Biscuits Tin)",
            "shortDesc": "Aneka biskuit renyah dalam kaleng ikonik (Assorted crispy biscuits in iconic tin)",
            "longDesc": "Koleksi biskuit klasik Khong Guan dengan berbagai rasa: vanila, cokelat, dan keju. Hadir dalam kaleng dekoratif yang bisa digunakan kembali. (Classic Khong Guan biscuit collection with vanilla, chocolate, and cheese varieties in a reusable decorative tin.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1558961363-fa8fdf82db35?w=800&q=80", "mimeType": "image/jpeg", "label": "Khong Guan Biscuit Tin" }]
          },
          "stockQuantity": 200,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Khong Guan", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 1600 },
              "appearance": { "color": "Red", "material": "Tin", "finish": "Glossy" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "Khong Guan Biscuit Factory Pte Ltd", "address": "Jakarta, DKI Jakarta, ID" },
              "commonOrGenericName": "Biskuit Assorted (Assorted Biscuits)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 1600 }
            }
          }
        },
        {
          "id": "res-biscuit-roma-kelapa-002",
          "descriptor": {
            "name": "Biskuit Roma Kelapa (Roma Coconut Biscuits)",
            "shortDesc": "Biskuit gurih rasa kelapa asli (Crunchy biscuits with authentic coconut flavor)",
            "longDesc": "Biskuit Roma Kelapa dibuat dari parutan kelapa pilihan dengan tekstur renyah yang khas. Cocok untuk teman minum teh atau kopi pagi hari. (Roma Coconut Biscuits made from selected coconut flakes with a distinctive crunchy texture, perfect with morning tea or coffee.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1506280754576-f6fa8a873550?w=800&q=80", "mimeType": "image/jpeg", "label": "Roma Coconut Biscuits" }]
          },
          "stockQuantity": 500,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Roma", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 200 },
              "appearance": { "color": "Yellow", "material": "Paper", "finish": "Matte" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "PT Mayora Indah Tbk", "address": "Tangerang, Banten, ID" },
              "commonOrGenericName": "Biskuit Kelapa (Coconut Biscuit)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 200 }
            }
          }
        },
        {
          "id": "res-chips-chitato-sapi-003",
          "descriptor": {
            "name": "Keripik Chitato Rasa Sapi Panggang (Chitato Beef BBQ Potato Chips)",
            "shortDesc": "Keripik kentang gurih rasa sapi panggang (Crunchy potato chips with BBQ beef flavor)",
            "longDesc": "Chitato hadir dengan rasa Sapi Panggang yang kaya dan bumbu khas Indonesia. Dibuat dari kentang pilihan yang diproses dengan teknologi modern untuk tekstur renyah sempurna. (Chitato with rich BBQ Beef flavor and distinctive Indonesian seasoning, made from selected potatoes for perfect crunch.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1621939514649-280e2ee25f60?w=800&q=80", "mimeType": "image/jpeg", "label": "Chitato BBQ Beef Chips" }]
          },
          "stockQuantity": 1000,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Chitato", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 68 },
              "appearance": { "color": "Red", "material": "Plastic", "finish": "Glossy" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "PT Indofood CBP Sukses Makmur Tbk", "address": "Jakarta, DKI Jakarta, ID" },
              "commonOrGenericName": "Keripik Kentang (Potato Chips)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 68 }
            }
          }
        },
        {
          "id": "res-chips-qtela-singkong-004",
          "descriptor": {
            "name": "Keripik Qtela Singkong Pedas (Qtela Spicy Cassava Chips)",
            "shortDesc": "Keripik singkong renyah dengan rasa pedas menggigit (Crunchy cassava chips with a fiery kick)",
            "longDesc": "Qtela Singkong Pedas dibuat dari singkong segar pilihan yang digoreng sempurna dengan bumbu pedas khas Indonesia. Teksturnya lebih tebal dan renyah dibanding keripik biasa. (Qtela Spicy Cassava Chips made from fresh selected cassava, perfectly fried with distinctive Indonesian spicy seasoning. Thicker and crunchier than regular chips.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1566478989037-eec170784d0b?w=800&q=80", "mimeType": "image/jpeg", "label": "Qtela Spicy Cassava Chips" }]
          },
          "stockQuantity": 800,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Qtela", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 110 },
              "appearance": { "color": "Orange", "material": "Plastic", "finish": "Glossy" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "PT Garudafood Putra Putri Jaya Tbk", "address": "Pati, Jawa Tengah, ID" },
              "commonOrGenericName": "Keripik Singkong (Cassava Chips)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 110 }
            }
          }
        },
        {
          "id": "res-cake-lapis-legit-005",
          "descriptor": {
            "name": "Kue Lapis Legit Premium (Premium Spekkoek Layered Cake)",
            "shortDesc": "Kue lapis klasik Belanda-Indonesia dengan lapisan sempurna (Classic Dutch-Indonesian layered cake with perfect layers)",
            "longDesc": "Lapis Legit adalah kue tradisional warisan Belanda-Indonesia yang dibuat dengan lebih dari 18 lapisan tipis. Setiap lapisan dipanggang satu per satu menggunakan rempah pilihan seperti kayu manis, cengkeh, dan kapulaga. (Lapis Legit is a traditional Dutch-Indonesian heritage cake with over 18 thin layers, each baked individually using selected spices like cinnamon, cloves, and cardamom.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=800&q=80", "mimeType": "image/jpeg", "label": "Lapis Legit Premium" }]
          },
          "stockQuantity": 50,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Dapur Lapis", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 700 },
              "appearance": { "color": "Brown", "material": "Paper Box", "finish": "Matte" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "CV Dapur Lapis Nusantara", "address": "Bandung, Jawa Barat, ID" },
              "commonOrGenericName": "Kue Lapis Legit (Spekkoek Cake)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 700 }
            }
          }
        },
        {
          "id": "res-cake-bolu-pandan-006",
          "descriptor": {
            "name": "Bolu Gulung Pandan (Pandan Swiss Roll Cake)",
            "shortDesc": "Bolu lembut gulung dengan krim pandan harum (Soft sponge roll with fragrant pandan cream)",
            "longDesc": "Bolu Gulung Pandan kami dibuat dari tepung pilihan dengan pewarna alami daun pandan segar. Diisi dengan krim butter pandan yang lembut dan tidak terlalu manis. (Our Pandan Swiss Roll Cake is made from select flour with natural coloring from fresh pandan leaves, filled with soft pandan butter cream.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1586985289688-ca3cf47d3e6e?w=800&q=80", "mimeType": "image/jpeg", "label": "Pandan Swiss Roll" }]
          },
          "stockQuantity": 75,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Bolu Nusantara", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 350 },
              "appearance": { "color": "Green", "material": "Cardboard", "finish": "Matte" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "UD Bolu Nusantara", "address": "Surabaya, Jawa Timur, ID" },
              "commonOrGenericName": "Bolu Gulung (Swiss Roll Cake)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 350 }
            }
          }
        },
        {
          "id": "res-biscuit-monde-butter-007",
          "descriptor": {
            "name": "Biskuit Monde Butter Cookies (Monde Butter Cookies)",
            "shortDesc": "Kue kering butter premium dalam kaleng cantik (Premium butter cookies in a beautiful tin)",
            "longDesc": "Monde Butter Cookies hadir dalam kaleng premium dengan pilihan cookies berbentuk bunga, pretzel, dan cinnamon. Dibuat dari butter pilihan berkualitas tinggi dengan tekstur yang lumer di mulut. (Monde Butter Cookies in premium tin with flower-shaped, pretzel, and cinnamon cookie varieties. Made from high-quality butter with a melt-in-your-mouth texture.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1548365328-8c6db3220e4c?w=800&q=80", "mimeType": "image/jpeg", "label": "Monde Butter Cookies Tin" }]
          },
          "stockQuantity": 150,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Monde", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 454 },
              "appearance": { "color": "Blue", "material": "Tin", "finish": "Glossy" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "PT Monde Mahkota Biskuit", "address": "Jakarta, DKI Jakarta, ID" },
              "commonOrGenericName": "Kue Kering Butter (Butter Cookies)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 454 }
            }
          }
        },
        {
          "id": "res-chips-taro-original-008",
          "descriptor": {
            "name": "Snack Taro Rasa Original (Taro Original Flavor Snack)",
            "shortDesc": "Snack tapioka renyah rasa original yang disukai semua usia (Crunchy tapioca snack loved by all ages)",
            "longDesc": "Taro Original adalah snack berbahan dasar tapioka dengan tekstur berlubang yang sangat renyah. Dibumbui dengan rasa original yang gurih dan ringan, cocok untuk camilan kapan saja. (Taro Original is a tapioca-based snack with a hollow, extra-crunchy texture. Seasoned with savory and light original flavor, perfect for snacking anytime.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1513558161293-cdaf765ed2fd?w=800&q=80", "mimeType": "image/jpeg", "label": "Taro Original Snack" }]
          },
          "stockQuantity": 600,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Taro", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 160 },
              "appearance": { "color": "Purple", "material": "Plastic", "finish": "Glossy" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "PT Garudafood Putra Putri Jaya Tbk", "address": "Pati, Jawa Tengah, ID" },
              "commonOrGenericName": "Snack Tapioka (Tapioca Snack)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 160 }
            }
          }
        },
        {
          "id": "res-cake-brownies-amanda-009",
          "descriptor": {
            "name": "Brownies Kukus Amanda Cokelat (Amanda Steamed Chocolate Brownies)",
            "shortDesc": "Brownies kukus lembut khas Bandung dengan cokelat premium (Soft steamed Bandung-style brownies with premium chocolate)",
            "longDesc": "Brownies Kukus Amanda adalah ikon kuliner Bandung yang terkenal. Dibuat dengan cokelat premium, teksturnya sangat lembut dan basah. Tersedia dalam varian original, keju, dan tiramisu. (Amanda Steamed Brownies are a famous Bandung culinary icon. Made with premium chocolate, extremely soft and moist. Available in original, cheese, and tiramisu variants.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1606313564200-e75d5e30476c?w=800&q=80", "mimeType": "image/jpeg", "label": "Amanda Steamed Brownies" }]
          },
          "stockQuantity": 60,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Amanda", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 800 },
              "appearance": { "color": "Brown", "material": "Cardboard", "finish": "Matte" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "CV Amanda Brownies", "address": "Bandung, Jawa Barat, ID" },
              "commonOrGenericName": "Brownies Kukus (Steamed Brownies)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 800 }
            }
          }
        },
        {
          "id": "res-biscuit-biskuat-energy-010",
          "descriptor": {
            "name": "Biskuit Biskuat Energy Cokelat (Biskuat Energy Chocolate Biscuits)",
            "shortDesc": "Biskuit energi cokelat bergizi tinggi untuk aktivitas sehari-hari (High-nutrition chocolate energy biscuits for daily activities)",
            "longDesc": "Biskuat Energy Cokelat mengandung susu dan 7 vitamin serta mineral esensial. Dirancang untuk memberikan energi optimal untuk anak-anak dan remaja yang aktif. (Biskuat Energy Chocolate contains milk and 7 essential vitamins and minerals. Designed to provide optimal energy for active children and teenagers.)",
            "mediaFile": [{ "uri": "https://images.unsplash.com/photo-1499195333224-3ce974eecb47?w=800&q=80", "mimeType": "image/jpeg", "label": "Biskuat Energy Chocolate" }]
          },
          "stockQuantity": 1200,
          "resourceAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailResource/v2.1/context.jsonld",
            "@type": "RetailResource",
            "identity": { "brand": "Biskuat", "originCountry": "ID" },
            "physical": {
              "weight": { "unitCode": "G", "unitQuantity": 120 },
              "appearance": { "color": "Brown", "material": "Plastic", "finish": "Glossy" }
            },
            "packagedGoodsDeclaration": {
              "manufacturerOrPacker": { "type": "MANUFACTURER", "name": "PT Kraft Foods Indonesia", "address": "Jakarta, DKI Jakarta, ID" },
              "commonOrGenericName": "Biskuit Energi (Energy Biscuit)",
              "netQuantity": { "unitCode": "G", "unitQuantity": 120 }
            }
          }
        }
      ],
      "offers": [
        {
          "id": "offer-biscuit-khong-guan-001",
          "descriptor": { "name": "Biskuit Kaleng Khong Guan (Khong Guan Assorted Biscuits Tin)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-biscuit-khong-guan-001"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-biscuit-khong-guan-001-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Biskuit Kaleng Khong Guan", "amount": 185000, "type": "BASE_PRICE" }],
              "totalAmount": 185000, "paymentMethods": ["PREPAID", "COD", "UPI"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP" },
              "cancellation": { "allowed": true, "window": "PT2H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP", "subjectToAvailability": true }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 20, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "09:00", "end": "21:00" } }]
            }
          }
        },
        {
          "id": "offer-biscuit-roma-kelapa-002",
          "descriptor": { "name": "Biskuit Roma Kelapa (Roma Coconut Biscuits)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-biscuit-roma-kelapa-002"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-biscuit-roma-kelapa-002-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Biskuit Roma Kelapa", "amount": 18500, "type": "BASE_PRICE" }],
              "totalAmount": 18500, "paymentMethods": ["PREPAID", "COD", "UPI"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP" },
              "cancellation": { "allowed": true, "window": "PT2H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP", "subjectToAvailability": true }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 20, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "09:00", "end": "21:00" } }]
            }
          }
        },
        {
          "id": "offer-chips-chitato-sapi-003",
          "descriptor": { "name": "Keripik Chitato Rasa Sapi Panggang (Chitato Beef BBQ Potato Chips)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-chips-chitato-sapi-003"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-chips-chitato-sapi-003-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Keripik Chitato Sapi Panggang", "amount": 12000, "type": "BASE_PRICE" }],
              "totalAmount": 12000, "paymentMethods": ["PREPAID", "COD", "UPI"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": false },
              "cancellation": { "allowed": true, "window": "PT1H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": true, "window": "P3D", "method": "SELLER_PICKUP", "subjectToAvailability": true }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 15, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "08:00", "end": "22:00" } }]
            }
          }
        },
        {
          "id": "offer-chips-qtela-singkong-004",
          "descriptor": { "name": "Keripik Qtela Singkong Pedas (Qtela Spicy Cassava Chips)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-chips-qtela-singkong-004"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-chips-qtela-singkong-004-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Keripik Qtela Singkong Pedas", "amount": 15000, "type": "BASE_PRICE" }],
              "totalAmount": 15000, "paymentMethods": ["PREPAID", "COD", "UPI"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": false },
              "cancellation": { "allowed": true, "window": "PT1H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": true, "window": "P3D", "method": "SELLER_PICKUP", "subjectToAvailability": true }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 15, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "08:00", "end": "22:00" } }]
            }
          }
        },
        {
          "id": "offer-cake-lapis-legit-005",
          "descriptor": { "name": "Kue Lapis Legit Premium (Premium Spekkoek Layered Cake)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-cake-lapis-legit-005"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-cake-lapis-legit-005-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Kue Lapis Legit Premium", "amount": 185000, "type": "BASE_PRICE" }],
              "totalAmount": 185000, "paymentMethods": ["PREPAID", "COD"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": false },
              "cancellation": { "allowed": true, "window": "PT4H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": false }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 10, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "09:00", "end": "18:00" } }]
            }
          }
        },
        {
          "id": "offer-cake-bolu-pandan-006",
          "descriptor": { "name": "Bolu Gulung Pandan (Pandan Swiss Roll Cake)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-cake-bolu-pandan-006"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-cake-bolu-pandan-006-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Bolu Gulung Pandan", "amount": 65000, "type": "BASE_PRICE" }],
              "totalAmount": 65000, "paymentMethods": ["PREPAID", "COD"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": false },
              "cancellation": { "allowed": true, "window": "PT2H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": false }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 15, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "08:00", "end": "20:00" } }]
            }
          }
        },
        {
          "id": "offer-biscuit-monde-butter-007",
          "descriptor": { "name": "Biskuit Monde Butter Cookies (Monde Butter Cookies)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-biscuit-monde-butter-007"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-biscuit-monde-butter-007-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Biskuit Monde Butter Cookies", "amount": 120000, "type": "BASE_PRICE" }],
              "totalAmount": 120000, "paymentMethods": ["PREPAID", "COD", "UPI"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP" },
              "cancellation": { "allowed": true, "window": "PT2H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP", "subjectToAvailability": true }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 20, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "09:00", "end": "21:00" } }]
            }
          }
        },
        {
          "id": "offer-chips-taro-original-008",
          "descriptor": { "name": "Snack Taro Rasa Original (Taro Original Flavor Snack)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-chips-taro-original-008"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-chips-taro-original-008-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Snack Taro Rasa Original", "amount": 22000, "type": "BASE_PRICE" }],
              "totalAmount": 22000, "paymentMethods": ["PREPAID", "COD", "UPI"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": false },
              "cancellation": { "allowed": true, "window": "PT1H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": true, "window": "P3D", "method": "SELLER_PICKUP", "subjectToAvailability": true }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 15, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "08:00", "end": "22:00" } }]
            }
          }
        },
        {
          "id": "offer-cake-brownies-amanda-009",
          "descriptor": { "name": "Brownies Kukus Amanda Cokelat (Amanda Steamed Chocolate Brownies)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-cake-brownies-amanda-009"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-cake-brownies-amanda-009-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Brownies Kukus Amanda Cokelat", "amount": 95000, "type": "BASE_PRICE" }],
              "totalAmount": 95000, "paymentMethods": ["PREPAID", "COD"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": false },
              "cancellation": { "allowed": true, "window": "PT4H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": false }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 10, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "09:00", "end": "18:00" } }]
            }
          }
        },
        {
          "id": "offer-biscuit-biskuat-energy-010",
          "descriptor": { "name": "Biskuit Biskuat Energy Cokelat (Biskuat Energy Chocolate Biscuits)" },
          "provider": { "id": "provider-toko-nusantara-001", "descriptor": { "name": "Toko Camilan Nusantara" } },
          "resourceIds": ["res-biscuit-biskuat-energy-010"],
          "validity": { "startDate": "2026-04-11T00:00:00Z", "endDate": "2027-04-11T00:00:00Z" },
          "considerations": [{
            "id": "offer-biscuit-biskuat-energy-010-price", "status": { "code": "ACTIVE", "name": "ACTIVE" },
            "considerationAttributes": {
              "@context": "https://schema.beckn.io/RetailConsideration/v2.1/context.jsonld",
              "@type": "RetailConsideration", "currency": "IDR",
              "breakup": [{ "title": "Biskuit Biskuat Energy Cokelat", "amount": 12500, "type": "BASE_PRICE" }],
              "totalAmount": 12500, "paymentMethods": ["PREPAID", "COD", "UPI"]
            }
          }],
          "offerAttributes": {
            "@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld",
            "@type": "RetailOffer",
            "policies": {
              "returns": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP" },
              "cancellation": { "allowed": true, "window": "PT2H", "cutoffEvent": "BEFORE_PACKING" },
              "replacement": { "allowed": true, "window": "P7D", "method": "SELLER_PICKUP", "subjectToAvailability": true }
            },
            "paymentConstraints": { "codAvailable": true },
            "serviceability": {
              "distanceConstraint": { "maxDistance": 20, "unit": "KM" },
              "timing": [{ "daysOfWeek": ["MON","TUE","WED","THU","FRI","SAT","SUN"], "timeRange": { "start": "09:00", "end": "21:00" } }]
            }
          }
        }
      ]
    }
  ]
}' | jq .
